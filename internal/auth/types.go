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

type registerRequest struct {
	FullName             string  `json:"full_name"`
	Email                string  `json:"email"`
	Password             string  `json:"password"`
	PasswordConfirmation string  `json:"password_confirmation"`
	DeviceName           *string `json:"device_name,omitempty"`
}

type loginRequest struct {
	Email      string  `json:"email"`
	Password   string  `json:"password"`
	DeviceName *string `json:"device_name,omitempty"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type authUserResponse struct {
	ID        string    `json:"id"`
	FullName  string    `json:"full_name"`
	Email     string    `json:"email"`
	Role      user.Role `json:"role"`
	AvatarKey *string   `json:"avatar_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type tokenPairResponse struct {
	AccessToken           string    `json:"access_token"`
	RefreshToken          string    `json:"refresh_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
}

type authResponse struct {
	User   authUserResponse  `json:"user"`
	Tokens tokenPairResponse `json:"tokens"`
}
