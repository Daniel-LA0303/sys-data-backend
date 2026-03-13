package user_dto

import "time"

// Request DTO
type CreateUserRequestDTO struct {
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

type UsersByOrganizationResponseDTO struct {
	OrgId    string `json:"orgId" db:"org_id"`
	UserId   string `json:"userId" db:"user_id"`
	Email    string `json:"email" db:"email"`
	Username string `json:"username" db:"username"`
	Role     string `json:"role" db:"role"`
}

type InviteUserResponseDTO struct {
	OrgId    string `json:"orgId" db:"org_id"`
	UserId   string `json:"userId" db:"user_id"`
	Email    string `json:"email" db:"email"`
	Username string `json:"username" db:"username"`
	Role     string `json:"role" db:"role"`
}

type LoginRequestDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=4"`
}

type LoginResponseDTO struct {
	Token    string `json:"token"`
	UserId   string `json:"userId" db:"user_id"`
	Email    string `json:"email" db:"email"`
	Username string `json:"username" db:"username"`
	OrgId    string `json:"orgId" db:"org_id"`
	OrgName  string `json:"orgName" db:"org_name"`
	RoleName string `json:"roleName" db:"role_name"`
	Theme    string `json:"theme" db:"theme_preference"`
	Language string `json:"language" db:"language"`
}

type CreateUserResponseDTO struct {
	Token    string `json:"token"`
	UserId   string `json:"userId" db:"user_id"`
	Email    string `json:"email" db:"email"`
	Username string `json:"username" db:"username"`
	OrgId    string `json:"orgId" db:"org_id"`
	OrgName  string `json:"orgName" db:"org_name"`
	RoleName string `json:"roleName" db:"role_name"`
}

type UpdateUserCustomSettingsDTO struct {
	UserId    string `json:"userId" db:"user_id" validate:"required"`
	FirstName string `json:"firstName" db:"first_name" validate:"required"`
	LastName  string `json:"lastName" db:"last_name" validate:"required"`
	Phone     string `json:"phone" db:"phone" validate:"required"`
	TimeZone  string `json:"timeZone" db:"time_zone" validate:"required"`
}

type RoleSmallRegisterInfo struct {
	RoleId   string `json:"roleId" db:"role_id"`
	RoleName string `json:"roleName" db:"role_name"`
}

type InviteUserRequestDTO struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=4"`
	OrgId    string `json:"orgId" validate:"required"`
	Role     string `json:"role" validate:"required"`
}

type UpdateLanguagePreferenceDTO struct {
	Language string `json:"language" validate:"required"`
}

type UpdateThemePreferenceDTO struct {
	Theme string `json:"theme" validate:"required"`
}
