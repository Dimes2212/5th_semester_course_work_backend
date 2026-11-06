package auth

import "errors"

var (
	// Регистрация
	ErrFullNameRequired    = errors.New("full name is required")
	ErrFullNameTooLong     = errors.New("full name is too long")
	ErrInvalidEmail        = errors.New("invalid email")
	ErrEmailTooLong        = errors.New("email is too long")
	ErrPasswordTooShort    = errors.New("password is too short")
	ErrPasswordTooLong     = errors.New("password is too long")
	ErrPasswordsDoNotMatch = errors.New("passwords do not match")
	ErrDeviceNameTooLong   = errors.New("device name is too long")
	ErrEmailAlreadyExists  = errors.New("email already exists")

	// Вход
	ErrInvalidCredentials = errors.New("invalid credentials")

	// Refresh-токен
	ErrRefreshTokenRequired = errors.New("refresh token is required")
	ErrInvalidRefreshToken  = errors.New("invalid refresh token")

	// Access-токен
	ErrAccessTokenRequired = errors.New("access token is required")
	ErrInvalidAccessToken  = errors.New("invalid access token")
)
