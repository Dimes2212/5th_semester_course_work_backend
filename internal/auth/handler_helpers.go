package auth

import (
	"errors"
	"net/http"

	"github.com/Dimes2212/5th_semester_course_work_backend/internal/apihelper"
)

func writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrEmailAlreadyExists):
		apihelper.WriteError(w, http.StatusConflict, "email_already_exists")

	case errors.Is(err, ErrInvalidCredentials):
		apihelper.WriteError(w, http.StatusUnauthorized, "invalid_credentials")

	case errors.Is(err, ErrInvalidRefreshToken):
		apihelper.WriteError(w, http.StatusUnauthorized, "invalid_refresh_token")

	case errors.Is(err, ErrRefreshTokenRequired):
		apihelper.WriteError(w, http.StatusBadRequest, "refresh_token_required")

	case isValidationError(err):
		apihelper.WriteError(w, http.StatusBadRequest, "validation_error")

	default:
		apihelper.WriteError(w, http.StatusInternalServerError, "internal_error")
	}
}

func isValidationError(err error) bool {
	return errors.Is(err, ErrFullNameRequired) ||
		errors.Is(err, ErrFullNameTooLong) ||
		errors.Is(err, ErrInvalidEmail) ||
		errors.Is(err, ErrEmailTooLong) ||
		errors.Is(err, ErrPasswordTooShort) ||
		errors.Is(err, ErrPasswordTooLong) ||
		errors.Is(err, ErrPasswordsDoNotMatch) ||
		errors.Is(err, ErrDeviceNameTooLong)
}

func makeAuthResponse(result AuthResult) authResponse {
	return authResponse{
		User: authUserResponse{
			ID:        result.User.ID,
			FullName:  result.User.FullName,
			Email:     result.User.Email,
			Role:      result.User.Role,
			AvatarKey: result.User.AvatarKey,
			CreatedAt: result.User.CreatedAt,
			UpdatedAt: result.User.UpdatedAt,
		},
		Tokens: tokenPairResponse{
			AccessToken:           result.Tokens.AccessToken,
			RefreshToken:          result.Tokens.RefreshToken,
			AccessTokenExpiresAt:  result.Tokens.AccessTokenExpiresAt,
			RefreshTokenExpiresAt: result.Tokens.RefreshTokenExpiresAt,
		},
	}
}
