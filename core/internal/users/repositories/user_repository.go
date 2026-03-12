package user_repository

import (
	"context"
	"database/sql"
	"errors"
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

type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

func (r *Repository) Create(ctx context.Context, db DBTX, u user_dto.CreateUserRequestDTO) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), 12)
	if err != nil {
		return err
	}

	query := `INSERT INTO users_tbl (username, email, password, created_at, updated_at) VALUES ($1,$2,$3,NOW(),NOW())`
	_, err = db.ExecContext(ctx, query, u.Username, u.Email, string(hashedPassword)) // db no r.db
	return err
}

// user_repository.go
func (r *Repository) GetByEmail(ctx context.Context, db DBTX, email string) (*user_dto.UserResponseDTO, error) {
	executor := db
	if executor == nil {
		executor = r.db
	}

	query := `SELECT user_id, username, email, created_at FROM users_tbl WHERE email = $1`

	var user user_dto.UserResponseDTO

	err := executor.GetContext(ctx, &user, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // no existe usuario, todo bien
		}
		log.Printf("ERROR in GetByEmail info: %v\n", err)
		return nil, err // error real
	}

	return &user, nil
}

func (r *Repository) GetUserByUsername(ctx context.Context, db DBTX, username string) (*user_dto.UserResponseDTO, error) {
	executor := db
	if executor == nil {
		executor = r.db
	}

	var user user_dto.UserResponseDTO
	query := `SELECT user_id, username, email, created_at FROM users_tbl WHERE username = $1`
	err := executor.GetContext(ctx, &user, query, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetInfoUserLoginAuth(ctx context.Context, email string) (*user_dto.LoginResponseDTO, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// we get info of user and organization
	query := `
	SELECT 
		ut.user_id,
		ut.email,
		ut.username,
		out2.org_id,
		oct.org_name
	FROM users_tbl ut
		inner  join organization_user_tbl out2 on out2.user_id = ut.user_id
	INNER JOIN organization_core_tbl oct 
		ON oct.org_id  = out2.org_id 
	WHERE ut.email = $1
	`

	var user user_dto.LoginResponseDTO
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		log.Printf("ERROR in GetInfoUserLoginAuth: %v\n", err)
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

func (r *Repository) GetInfoRoleByName(ctx context.Context, rolename string) (*user_dto.RoleSmallRegisterInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `SELECT role_id, role_name FROM user_roles_core_tbl WHERE role_name = $1`

	var role user_dto.RoleSmallRegisterInfo
	err := r.db.GetContext(ctx, &role, query, rolename)
	if err != nil {
		return nil, err
	}

	return &role, nil
}

func (r *Repository) AssignRole(
	ctx context.Context,
	db DBTX,
	roleId string,
	userId string,
	orgId string,
) error {

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
		INSERT INTO user_role_assignment_tbl
			(user_id, role_id, org_id, assigned_at)
		VALUES ($1,$2,$3,NOW())
	`

	_, err := db.ExecContext(ctx, query, userId, roleId, orgId)
	if err != nil {
		log.Printf("ERROR in user_role_assignment_tbl insert: %v\n", err)
		return err
	}

	return nil
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

func (r *Repository) GetUsersByOrganizationPaginated(
	ctx context.Context,
	limit, offset int,
	orgId string,
) ([]user_dto.UsersByOrganizationResponseDTO, int, error) {

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var users []user_dto.UsersByOrganizationResponseDTO
	var total int

	// 1 paginated
	query := `
		SELECT out2.org_id, out2.user_id, out2.role, ut.username, ut.email
		FROM organization_user_tbl out2
		INNER JOIN users_tbl ut ON out2.user_id = ut.user_id
		WHERE out2.org_id = $3
		ORDER BY ut.created_at DESC
		LIMIT $1 OFFSET $2
	`

	err := r.db.SelectContext(ctx, &users, query, limit, offset, orgId)
	if err != nil {
		return nil, 0, err
	}

	// 2 total users
	countQuery := `
		SELECT COUNT(*)
		FROM organization_user_tbl
		WHERE org_id = $1
	`

	err = r.db.GetContext(ctx, &total, countQuery, orgId)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// get users by organization paginated
// TODO this repo isn't paginated yet
func (r *Repository) GetUsersByOrganizationChatPaginated(
	ctx context.Context,
	orgId string,
) ([]user_dto.UsersByOrganizationResponseDTO, error) {

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var users []user_dto.UsersByOrganizationResponseDTO

	query := `
		SELECT out2.org_id, out2.user_id, out2.role, ut.username, ut.email
		FROM organization_user_tbl out2
		INNER JOIN users_tbl ut ON out2.user_id = ut.user_id
		WHERE out2.org_id = $1
		ORDER BY ut.created_at DESC
	`

	err := r.db.SelectContext(ctx, &users, query, orgId)
	if err != nil {
		log.Printf("ERROR in GetUsersByOrganizationChat insert: %v\n", err)
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
		log.Printf("INSERT ERROR in CreateCustomSettings: %v\n", err)
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
