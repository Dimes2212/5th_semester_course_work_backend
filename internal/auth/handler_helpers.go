package auth

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, errorResponse{Error: code})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain a single JSON object")
	}

	return nil
}

func writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrEmailAlreadyExists):
		writeError(w, http.StatusConflict, "email_already_exists")

	case errors.Is(err, ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "invalid_credentials")

	case errors.Is(err, ErrInvalidRefreshToken):
		writeError(w, http.StatusUnauthorized, "invalid_refresh_token")

	case errors.Is(err, ErrRefreshTokenRequired):
		writeError(w, http.StatusBadRequest, "refresh_token_required")

	case isValidationError(err):
		writeError(w, http.StatusBadRequest, "validation_error")

	default:
		writeError(w, http.StatusInternalServerError, "internal_error")
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
