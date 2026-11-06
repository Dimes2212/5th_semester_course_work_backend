package auth

import (
	"time"

	"github.com/Dimes2212/5th_semester_course_work_backend/internal/user"
	"github.com/golang-jwt/jwt/v5"
)

type RefreshSession struct {
	ID         string
	UserID     string
	TokenHash  string
	DeviceName *string
	ExpiresAt  time.Time
	CreatedAt  time.Time
	RevokedAt  *time.Time
}

type RegisterInput struct {
	FullName             string
	Email                string
	Password             string
	PasswordConfirmation string
	DeviceName           *string
}

type LoginInput struct {
	Email      string
	Password   string
	DeviceName *string
}

type TokenPair struct {
	AccessToken           string
	RefreshToken          string
	AccessTokenExpiresAt  time.Time
	RefreshTokenExpiresAt time.Time
}

type AuthUser struct {
	ID        string
	FullName  string
	Email     string
	Role      user.Role
	AvatarKey *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AuthResult struct {
	User   AuthUser
	Tokens TokenPair
}

type AccessTokenClaims struct {
	UserID string    `json:"user_id"`
	Role   user.Role `json:"role"`
	jwt.RegisteredClaims
}

type Identity struct {
	UserID string
	Role   user.Role
}
