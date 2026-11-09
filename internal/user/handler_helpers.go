package user

import (
	"errors"
	"net/http"

	"github.com/Dimes2212/5th_semester_course_work_backend/internal/apihelper"
)

func writeUserError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrUserNotFound):
		apihelper.WriteError(w, http.StatusUnauthorized, "unauthorized")

	case errors.Is(err, ErrFullNameRequired),
		errors.Is(err, ErrFullNameTooLong):
		apihelper.WriteError(w, http.StatusBadRequest, "validation_error")

	default:
		apihelper.WriteError(w, http.StatusInternalServerError, "internal_error")
	}
}

func makeProfileResponse(profile Profile) profileResponse {
	return profileResponse{
		ID:        profile.ID,
		FullName:  profile.FullName,
		Email:     profile.Email,
		Role:      profile.Role,
		AvatarKey: profile.AvatarKey,
		CreatedAt: profile.CreatedAt,
		UpdatedAt: profile.UpdatedAt,
	}
}
