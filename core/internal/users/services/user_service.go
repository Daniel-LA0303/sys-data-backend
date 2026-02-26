package users_service

import (
	"context"
	"core/project/core/common/errors"
	organization_dto "core/project/core/internal/organization/dtos"
	organization_repository "core/project/core/internal/organization/repositories"
	"core/project/core/internal/users/auth"
	user_dto "core/project/core/internal/users/dtos"
	user_repository "core/project/core/internal/users/repositories"
	"database/sql"
	stdErrors "errors"
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

// this function does not have validation
func (s *Service) RegisterUser(ctx context.Context, input user_dto.CreateUserDTO) (*user_dto.AuthResponseDTO, error) {

	// 1. valid if user with email already exists
	existing, _ := s.repo.GetByEmail(ctx, s.db, input.Email)
	if existing != nil && existing.UserID != "" {
		return nil, errors.NewConflictError("User already exists")
	}

	// 2. init transaction
	tx, err := s.tx.BeginTxx(ctx, nil)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}

	// 2.1. defer rollback
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	// 3. create new user
	if err = s.repo.Create(ctx, tx, input); err != nil {
		return nil, err
	}

	// 4. get user that was just created
	user, err := s.repo.GetByEmail(ctx, tx, input.Email)
	if err != nil {
		return nil, err
	}

	// 5. create default org
	orgSlug := "default-" + user.UserID[:6]

	// 5.1 build default org
	org := organization_dto.OrganizationInfoSmallDTO{
		OrgName:     "default",
		OrgSlug:     orgSlug,
		OwnerUserId: user.UserID,
		Status:      true,
	}

	// 5.2 create default org
	orgID, err := s.orgRepo.CreateOrganization(ctx, tx, org)
	if err != nil {
		return nil, err
	}

	// 4. get info of role
	roleInfo, err := s.repo.GetInfoRoleByName(ctx, "ROLE_ADMIN")

	if err = s.repo.AssignRole(ctx, roleInfo.RoleId, user.UserID, orgID); err != nil {
		return nil, err
	}

	// 6. commit, this is where the transaction ends
	if err = tx.Commit(); err != nil {
		return nil, errors.NewDatabaseError(err)
	}

	// 7. we create the JWT
	token, err := auth.CreateJWT(user.UserID, orgID, roleInfo.RoleName, user.Email)
	if err != nil {
		return nil, err
	}

	// 8. return response
	return &user_dto.AuthResponseDTO{
		Token:    token,
		UserId:   user.UserID,
		Email:    user.Email,
		Username: user.Username,
		OrgName:  org.OrgName,
		OrgId:    orgID,
		RoleName: roleInfo.RoleName,
	}, nil
}

func (s *Service) Login(ctx context.Context, input user_dto.LoginDTO) (*user_dto.AuthResponseDTO, error) {

	// 1. validate email and password
	if input.Email == "" || input.Password == "" {
		return nil, errors.NewValidationError("Email and password are required")
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
	token, err := auth.CreateJWT(userID, userInfo.OrgId, roleInfo.RoleName, userInfo.Email)
	if err != nil {
		return nil, errors.NewInternalError(err)
	}

	// 6. attach token
	userInfo.Token = token
	userInfo.RoleName = roleInfo.RoleName

	return userInfo, nil
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

func (s *Service) UpsertUserCustomSettings(
	ctx context.Context,
	input user_dto.UpdateUserCustomSettingsDTO,
) error {

	// 1. get id from context and valid is the same with token and sended, is most secure

	userID := auth.GetUserIDFromContext(ctx)

	log.Printf("userid from token conetxt is: %v\n", userID)
	log.Printf("userid from request is: %v\n", input.UserId)
	if userID == "" || userID != input.UserId {
		return errors.NewUnauthorizedError("Unauthorized")
	}

	input.UserId = userID

	// 2. check if custom settings user exists
	existing, err := s.repo.GetCustomSettingsByUserID(ctx, userID)

	if err != nil {
		if stdErrors.Is(err, sql.ErrNoRows) {
			// 3. if does not exists, then we create it
			log.Println("NO ROWS FOUND → creating settings")
			return s.repo.CreateCustomSettings(ctx, input)
		}
		log.Printf("DB ERROR: %v\n", err)
		return errors.NewDatabaseError(err)
	}

	// 4. if exists, then we update it
	_ = existing
	return s.repo.UpdateInfoUserCustomSettings(ctx, input)
}
