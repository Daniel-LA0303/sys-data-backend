package user_dto

import "time"

// Request DTO
type CreateUserDTO struct {
	Username string
	Email    string
	Password string
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
	Token string `json:"token"`
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
