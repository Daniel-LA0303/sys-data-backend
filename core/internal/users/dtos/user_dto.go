package user_dto

import "time"

// Request DTO
type CreateUserDTO struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=4"`
}

// Response DTO
type UserResponseDTO struct {
	UserID    string    `json:"user_id" db:"user_id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type LoginDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponseDTO struct {
	Token    string `json:"token"`
	UserId   string `json:"userId" db:"user_id"`
	Email    string `json:"email" db:"email"`
	Username string `json:"username" db:"username"`
	OrgId    string `json:"orgId" db:"org_id"`
	OrgName  string `json:"orgName" db:"org_name"`
	RoleName string `json:"roleName" db:"role_name"`
}

type UpdateUserCustomSettingsDTO struct {
	UserId           string `json:"userId" db:"user_id"`
	FirstName        string `json:"firstName" db:"first_name"`
	LastName         string `json:"lastName" db:"last_name"`
	Phone            string `json:"phone" db:"phone"`
	AvatarUrl        string `json:"avatarUrl" db:"avatar_url"`
	TimeZone         string `json:"timeZone" db:"time_zone"`
	Language         string `json:"language" db:"language"`
	ThemePreference  string `json:"themePreference" db:"theme_preference"`
	ProfileCompleted bool   `json:"profileCompleted" db:"profile_completed"`
}

type RoleSmallRegisterInfo struct {
	RoleId   string `json:"roleId" db:"role_id"`
	RoleName string `json:"roleName" db:"role_name"`
}
