package user

import "time"

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type User struct {
	ID           string
	FullName     string
	Email        string
	PasswordHash string
	Role         Role
	AvatarKey    *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

type Profile struct {
	ID        string
	FullName  string
	Email     string
	Role      Role
	AvatarKey *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UpdateProfileInput struct {
	FullName string
}

type updateProfileRequest struct {
	FullName string `json:"full_name"`
}

type profileResponse struct {
	ID        string    `json:"id"`
	FullName  string    `json:"full_name"`
	Email     string    `json:"email"`
	Role      Role      `json:"role"`
	AvatarKey *string   `json:"avatar_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
