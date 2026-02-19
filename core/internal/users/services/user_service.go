package users_service

import (
	"context"
	"core/project/core/common/errors"
	"core/project/core/internal/users/auth"
	user_dto "core/project/core/internal/users/dtos"
	user_repository "core/project/core/internal/users/repositories"
	"database/sql"
	stdErrors "errors"
	"log"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo *user_repository.Repository
}

func NewService(repo *user_repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) RegisterUser(ctx context.Context, input user_dto.CreateUserDTO) (string, error) {

	if input.Email == "" {
		return "", errors.NewValidationError("Email is required")
	}
	if len(input.Password) < 4 {
		return "", errors.NewValidationError("Password must be at least 8 characters")
	}

	existing, err := s.repo.GetByEmail(ctx, input.Email)
	if err != nil && err != sql.ErrNoRows {
		return "", errors.NewDatabaseError(err)
	}
	if existing != nil {
		return "", errors.NewConflictError("User with this email already exists")
	}

	existingUsername, _ := s.repo.GetByUsername(ctx, input.Username)
	if existingUsername != nil {
		return "", errors.NewConflictError("User with this username already exists")
	}

	err = s.repo.Create(ctx, input)
	if err != nil {
		return "", errors.NewDatabaseError(err)
	}

	// Obtener usuario recién creado
	user, err := s.repo.GetByEmail(ctx, input.Email)
	if err != nil {
		return "", errors.NewDatabaseError(err)
	}

	token, err := auth.CreateJWT(user.UserID)
	if err != nil {
		return "", err
	}

	return token, nil
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

	// 4. create token
	token, err := auth.CreateJWT(userID)
	if err != nil {
		return nil, errors.NewInternalError(err)
	}

	// 5. get full user info (org, username, etc.)
	userInfo, err := s.repo.GetInfoUserLoginAuth(ctx, input.Email)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}

	// 6. attach token
	userInfo.Token = token

	return userInfo, nil
}

func (s *Service) GetUserByEmail(ctx context.Context, email string) (*user_dto.UserResponseDTO, error) {
	if email == "" {
		return nil, errors.NewValidationError("Email is required")
	}

	user, err := s.repo.GetByEmail(ctx, email)
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
