package user_repository

import (
	"context"
	"log"
	"time"

	// IMPORTANTE: Aquí importas la carpeta física (dtos)
	user_dto "core/project/core/internal/users/dtos"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Fíjate en el prefijo: user_dto.NombreDelStruct
func (r *Repository) Create(ctx context.Context, u user_dto.CreateUserDTO) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), 12)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO users_tbl (username, email, password, created_at, updated_at)
		VALUES ($1,$2,$3,NOW(),NOW())
	`
	_, err = r.db.ExecContext(ctx, query, u.Username, u.Email, string(hashedPassword))
	return err
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (*user_dto.UserResponseDTO, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `SELECT user_id, username, email, created_at FROM users_tbl WHERE email = $1`

	var user user_dto.UserResponseDTO
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		log.Printf("get emial ERROR: %v\n", err)
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *Repository) GetInfoUserLoginAuth(ctx context.Context, email string) (*user_dto.AuthResponseDTO, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
	SELECT 
		ut.user_id,
		ut.email,
		ut.username,
		oct.org_id,
		oct.org_name
	FROM users_tbl ut
	INNER JOIN organization_core_tbl oct 
		ON oct.owner_user_id = ut.user_id
	WHERE ut.email = $1
	`

	var user user_dto.AuthResponseDTO
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		log.Printf("get email auth ERROR: %v\n", err)
	}
	if err != nil {

		return nil, err
	}

	return &user, nil
}

func (r *Repository) GetByUsername(ctx context.Context, username string) (*user_dto.UserResponseDTO, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `SELECT user_id, username, email, created_at FROM users_tbl WHERE username = $1`

	var user user_dto.UserResponseDTO
	err := r.db.GetContext(ctx, &user, query, username)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *Repository) GetById(ctx context.Context, id string) (*user_dto.UserResponseDTO, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `SELECT user_id, username, email, created_at FROM users_tbl WHERE user_id = $1`

	var user user_dto.UserResponseDTO
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *Repository) GetPaginated(ctx context.Context, limit, offset int) ([]user_dto.UserResponseDTO, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// El query con LIMIT y OFFSET
	query := `
		SELECT user_id, username, email, created_at 
		FROM users_tbl 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2`

	var users []user_dto.UserResponseDTO
	// Usamos SelectContext para traer múltiples filas
	err := r.db.SelectContext(ctx, &users, query, limit, offset)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *Repository) GetAuthByEmail(ctx context.Context, email string) (string, string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
		SELECT user_id, password 
		FROM users_tbl 
		WHERE email = $1
	`

	var userID string
	var hashedPassword string

	err := r.db.QueryRowContext(ctx, query, email).Scan(&userID, &hashedPassword)
	if err != nil {
		return "", "", err
	}

	return userID, hashedPassword, nil
}

func (r *Repository) CreateCustomSettings(
	ctx context.Context,
	u user_dto.UpdateUserCustomSettingsDTO,
) error {

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
		INSERT INTO user_custom_settings_core_tbl
		(user_id, first_name, last_name, phone, avatar_url, time_zone, language, theme_preference, profile_completed)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		u.UserId,
		u.FirstName,
		u.LastName,
		u.Phone,
		u.AvatarUrl,
		u.TimeZone,
		u.Language,
		u.ThemePreference,
		u.ProfileCompleted,
	)

	if err != nil {
		log.Printf("INSERT ERROR: %v\n", err)
	}

	return err
}

func (r *Repository) UpdateInfoUserCustomSettings(
	ctx context.Context,
	u user_dto.UpdateUserCustomSettingsDTO,
) error {

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
		UPDATE user_custom_settings_core_tbl
		SET 
			first_name = $1,
			last_name = $2,
			phone = $3,
			avatar_url = $4,
			time_zone = $5,
			language = $6,
			theme_preference = $7,
			profile_completed = $8
		WHERE user_id = $9
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		u.FirstName,
		u.LastName,
		u.Phone,
		u.AvatarUrl,
		u.TimeZone,
		u.Language,
		u.ThemePreference,
		u.ProfileCompleted,
		u.UserId,
	)

	if err != nil {
		log.Printf("UPDTAE ERROR: %v\n", err)
	}

	return err
}

func (r *Repository) GetCustomSettingsByUserID(
	ctx context.Context,
	userID string,
) (*user_dto.UpdateUserCustomSettingsDTO, error) {

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
		SELECT 
			user_id,
			first_name,
			last_name,
			phone,
			avatar_url,
			time_zone,
			language,
			theme_preference,
			profile_completed
		FROM user_custom_settings_core_tbl
		WHERE user_id = $1
	`

	var settings user_dto.UpdateUserCustomSettingsDTO

	err := r.db.GetContext(ctx, &settings, query, userID)
	if err != nil {
		log.Printf("GET ERROR: %v\n", err)
		return nil, err
	}

	return &settings, nil
}
