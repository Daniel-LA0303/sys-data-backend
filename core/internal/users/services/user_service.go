package users_service

import (
	"context"
	"core/project/core/common/errors"
	"core/project/core/common/validator"
	organization_dto "core/project/core/internal/organization/dtos"
	organization_repository "core/project/core/internal/organization/repositories"
	"core/project/core/internal/users/auth"
	user_dto "core/project/core/internal/users/dtos"
	user_repository "core/project/core/internal/users/repositories"
	"database/sql"
	"log"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo    *user_repository.Repository
	orgRepo *organization_repository.Repository
	tx      TxManager
	db      *sqlx.DB
}

type TxManager interface {
	BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error)
}

// Cambia esto en users_service.go
func NewService(repo *user_repository.Repository, orgRepo *organization_repository.Repository, db *sqlx.DB) *Service {
	return &Service{
		repo:    repo,
		orgRepo: orgRepo,
		tx:      db, // Aquí db actúa como TxManager
		db:      db, // Aquí db actúa como la conexión para queries simples
	}
}

// REGISTER USER WITH ROLE ADMIN
func (s *Service) RegisterUser(ctx context.Context, input user_dto.CreateUserRequestDTO) (*user_dto.CreateUserResponseDTO, error) {

	// 1, valid input
	validationErrors := validator.ValidateStruct(input)
	if validationErrors != nil {
		return nil, errors.NewValidationFiledsError("Invalid input", validationErrors)
	}

	// 2. valid if user with email already exists
	existing, _ := s.repo.GetByEmail(ctx, s.db, input.Email)
	if existing != nil && existing.UserID != "" {
		return nil, errors.NewConflictError("User already exists")
	}

	// 3. init transaction
	tx, err := s.tx.BeginTxx(ctx, nil)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}

	// 3.1. defer rollback
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	// 4. create new user
	if err = s.repo.Create(ctx, tx, input); err != nil {
		return nil, err
	}

	// 5. get user that was just created
	user, err := s.repo.GetByEmail(ctx, tx, input.Email)
	if err != nil {
		return nil, err
	}

	// 6. create default org
	orgSlug := "default-" + user.UserID[:6]

	// 6.1 build default org
	org := organization_dto.OrganizationInfoSmallDTO{
		OrgName:     "default",
		OrgSlug:     orgSlug,
		OwnerUserId: user.UserID,
		Status:      true,
	}

	// 6.2 create default org
	orgID, err := s.orgRepo.CreateOrganization(ctx, tx, org)
	if err != nil {
		return nil, err
	}

	// 7. get info of role admin
	roleInfo, err := s.repo.GetInfoRoleByName(ctx, "ROLE_ADMIN")

	// 8. assign role
	if err = s.repo.AssignRole(ctx, tx, roleInfo.RoleId, user.UserID, orgID); err != nil {
		return nil, err
	}

	// 9. insert data in organization_user_tbl
	orgUser := organization_dto.InsertUserOrganizationDTO{
		OrgId:  orgID,
		UserId: user.UserID,
		Role:   roleInfo.RoleName,
	}
	if err = s.orgRepo.InsertUserOrganization(ctx, tx, orgUser); err != nil {
		return nil, err
	}

	if err = s.repo.InsertUserSettingsDefault(ctx, tx, user.UserID); err != nil {
		return nil, err
	}

	// 10. commit, this is where the transaction ends
	if err = tx.Commit(); err != nil {
		return nil, errors.NewDatabaseError(err)
	}

	// 11. we create the JWT
	token, err := auth.CreateJWT(user.UserID, orgID, roleInfo.RoleName, user.Email, user.Username)
	if err != nil {
		return nil, err
	}

	// 12. return response
	return &user_dto.CreateUserResponseDTO{
		Token:    token,
		UserId:   user.UserID,
		Email:    user.Email,
		Username: user.Username,
		OrgName:  org.OrgName,
		OrgId:    orgID,
		RoleName: roleInfo.RoleName,
	}, nil
}

// LOGIN USER
func (s *Service) Login(ctx context.Context, input user_dto.LoginRequestDTO) (*user_dto.LoginResponseDTO, error) {

	// 1. validate email and password
	validationErrors := validator.ValidateStruct(input)
	if validationErrors != nil {
		return nil, errors.NewValidationFiledsError("Invalid input", validationErrors)
	}

	// 2. get auth credentials
	userID, hashedPassword, err := s.repo.GetAuthByEmail(ctx, input.Email)
	if err == sql.ErrNoRows {
		return nil, errors.NewUnauthorizedError("Email not found")
	}
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}

	// 3. validate password
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(input.Password))
	if err != nil {
		return nil, errors.NewUnauthorizedError("Invalid password")
	}

	// 4. get info of role
	roleInfo, err := s.repo.GetInfoRoleByName(ctx, "ROLE_ADMIN")

	// 5. get full user info (org, username, etc.)
	userInfo, err := s.repo.GetInfoUserLoginAuth(ctx, input.Email)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}

	// 4. create token
	token, err := auth.CreateJWT(userID, userInfo.OrgId, roleInfo.RoleName, userInfo.Email, userInfo.Username)
	if err != nil {
		return nil, errors.NewInternalError(err)
	}

	// 6. attach token
	userInfo.Token = token
	userInfo.RoleName = roleInfo.RoleName

	return userInfo, nil
}

// INVITE USER WITH ROLE MEMBER
func (s *Service) InviteUser(ctx context.Context, input user_dto.InviteUserRequestDTO) (user_dto.InviteUserResponseDTO, error) {

	// 1. validate input
	validationErrors := validator.ValidateStruct(input)
	if validationErrors != nil {
		return user_dto.InviteUserResponseDTO{}, errors.NewValidationFiledsError("Invalid input", validationErrors)
	}

	// 2. validate if user with email already exists
	existing, _ := s.repo.GetByEmail(ctx, s.db, input.Email)
	if existing != nil && existing.UserID != "" {
		return user_dto.InviteUserResponseDTO{}, errors.NewConflictError("User with this email already exists")
	}

	// 2.2 validate if user with username already exists
	existing2, err := s.repo.GetUserByUsername(ctx, s.db, input.Username)
	if err != nil {
		return user_dto.InviteUserResponseDTO{}, errors.NewDatabaseError(err)
	}
	if existing2 != nil {
		return user_dto.InviteUserResponseDTO{}, errors.NewConflictError("User with this username already exists")
	}

	log.Printf("valid")

	// 3. init transaction
	tx, err := s.tx.BeginTxx(ctx, nil)
	if err != nil {
		return user_dto.InviteUserResponseDTO{}, errors.NewDatabaseError(err)
	}

	// 3.1 defer rollback
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	userCreated := user_dto.CreateUserRequestDTO{
		Username: input.Username,
		Email:    input.Email,
		Password: input.Password,
	}

	// 4. create user
	if err = s.repo.Create(ctx, tx, userCreated); err != nil {
		return user_dto.InviteUserResponseDTO{}, err
	}
	log.Printf("created user")

	// 5. get created user
	user, err := s.repo.GetByEmail(ctx, tx, input.Email)
	if err != nil {
		return user_dto.InviteUserResponseDTO{}, err
	}
	log.Printf("get by email")

	// 6. get role member
	roleInfo, err := s.repo.GetInfoRoleByName(ctx, input.Role)
	if err != nil {
		return user_dto.InviteUserResponseDTO{}, err
	}
	log.Printf("get role info")

	// 7. assign role
	if err = s.repo.AssignRole(ctx, tx, roleInfo.RoleId, user.UserID, input.OrgId); err != nil {
		return user_dto.InviteUserResponseDTO{}, err
	}
	log.Printf("assign role")

	// 8. insert in organization_user_tbl
	orgUser := organization_dto.InsertUserOrganizationDTO{
		OrgId:  input.OrgId,
		UserId: user.UserID,
		Role:   roleInfo.RoleName,
	}

	if err = s.orgRepo.InsertUserOrganization(ctx, tx, orgUser); err != nil {
		return user_dto.InviteUserResponseDTO{}, err
	}

	if err = s.repo.InsertUserSettingsDefault(ctx, tx, user.UserID); err != nil {
		return user_dto.InviteUserResponseDTO{}, err
	}
	log.Printf("insert in organization_user_tbl")

	// 9. build response
	res := user_dto.InviteUserResponseDTO{
		UserId:   user.UserID,
		Email:    user.Email,
		Username: user.Username,
		Role:     roleInfo.RoleName,
		OrgId:    input.OrgId,
	}
	log.Printf("build response")

	// 9. commit
	if err = tx.Commit(); err != nil {
		return user_dto.InviteUserResponseDTO{}, errors.NewDatabaseError(err)
	}

	return res, nil
}

func (s *Service) GetUserByEmail(ctx context.Context, email string) (*user_dto.UserResponseDTO, error) {
	if email == "" {
		return nil, errors.NewValidationError("Email is required")
	}

	user, err := s.repo.GetByEmail(ctx, s.db, email) // agregar s.db
	if err == sql.ErrNoRows {
		return nil, errors.NewNotFoundError("User")
	}
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}

	return user, nil
}

func (s *Service) GetUsers(ctx context.Context, page, limit int) ([]user_dto.UserResponseDTO, error) {
	// Validaciones básicas
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	} // Tope de 100 por seguridad

	offset := (page - 1) * limit

	return s.repo.GetPaginated(ctx, limit, offset)
}

func (s *Service) GetUsersByOrganizationPaginated(
	ctx context.Context,
	page, limit int,
	orgId string,
) ([]user_dto.UsersByOrganizationResponseDTO, int, error) {

	if page < 1 {
		page = 1
	}

	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	return s.repo.GetUsersByOrganizationPaginated(ctx, limit, offset, orgId)
}

// get users by organization paginated
func (s *Service) GetUsersByOrganizationChatPaginated(
	ctx context.Context,
	orgId string,
) ([]user_dto.UsersByOrganizationResponseDTO, error) {

	return s.repo.GetUsersByOrganizationChatPaginated(ctx, orgId)
}

func (s *Service) UpdateLanguagePreference(
	ctx context.Context,
	language string,
) (user_dto.UpdateLanguagePreferenceDTO, error) { // Retornamos DTO y error
	userID := auth.GetUserIDFromContext(ctx)
	if userID == "" {
		return user_dto.UpdateLanguagePreferenceDTO{}, errors.NewUnauthorizedError("Unauthorized")
	}

	response := user_dto.UpdateLanguagePreferenceDTO{
		Language: language,
	}

	// 1. Obtener configuraciones existentes
	existing, err := s.repo.GetCustomSettingsByUserIdOptional(ctx, userID)
	if err != nil {
		return user_dto.UpdateLanguagePreferenceDTO{}, errors.NewDatabaseError(err)
	}

	// 2. Si no existe, crear registro por defecto con el lenguaje seleccionado
	if existing == nil {
		err = s.repo.InsertUserSettingsDefaultWithLanguage(ctx, userID, language)
		if err != nil {
			return user_dto.UpdateLanguagePreferenceDTO{}, errors.NewDatabaseError(err)
		}
		return response, nil
	}

	// 3. Si ya existe, actualizar solo el lenguaje
	err = s.repo.UpdateLanguagePreference(ctx, language, userID)
	if err != nil {
		return user_dto.UpdateLanguagePreferenceDTO{}, errors.NewDatabaseError(err)
	}

	return response, nil
}

func (s *Service) UpdateThemePreference(
	ctx context.Context,
	theme string,
) (user_dto.UpdateThemePreferenceDTO, error) {
	userID := auth.GetUserIDFromContext(ctx)
	if userID == "" {
		return user_dto.UpdateThemePreferenceDTO{}, errors.NewUnauthorizedError("Unauthorized")
	}

	response := user_dto.UpdateThemePreferenceDTO{
		Theme: theme,
	}

	// 1. get existing settings
	existing, err := s.repo.GetCustomSettingsByUserIdOptional(ctx, userID)
	if err != nil {
		return user_dto.UpdateThemePreferenceDTO{}, errors.NewDatabaseError(err)
	}

	// 2.if not exists, create with language
	if existing == nil {
		// Puedes crear un método específico que inserte con un lenguaje inicial
		return user_dto.UpdateThemePreferenceDTO{}, s.repo.InsertUserSettingsDefaultWithTheme(ctx, userID, theme)
	}

	// 3. if exists, update language
	return response, s.repo.UpdateThemePreference(ctx, theme, userID)
}

func (s *Service) CreateOrUpdateUserCustomSettings(
	ctx context.Context,
	input user_dto.UpdateUserCustomSettingsDTO,
) error {
	userID := auth.GetUserIDFromContext(ctx)

	// 1. valid used id
	if userID == "" || userID != input.UserId {
		return errors.NewUnauthorizedError("Unauthorized")
	}

	// 2. get existing settings
	existing, err := s.repo.GetCustomSettingsByUserID(ctx, userID)
	if err != nil {
		return errors.NewDatabaseError(err)
	}

	// 2. if not exists, create settings else update
	if existing == nil {
		return s.repo.CreateCustomSettings(ctx, input)
	}

	return s.repo.UpdateInfoUserCustomSettings(ctx, input)
}
