package user

import (
	"net/http"

	"github.com/Dimes2212/5th_semester_course_work_backend/internal/apihelper"
	"github.com/Dimes2212/5th_semester_course_work_backend/internal/requestctx"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := requestctx.CurrentUserFromContext(r.Context())
	if !ok {
		apihelper.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	profile, err := h.service.GetByID(r.Context(), currentUser.UserID)
	if err != nil {
		writeUserError(w, err)
		return
	}

	apihelper.WriteJSON(w, http.StatusOK, makeProfileResponse(profile))
}

func (h *Handler) UpdateName(w http.ResponseWriter, r *http.Request) {

	var request updateProfileRequest

	if err := apihelper.DecodeJSON(w, r, &request); err != nil {
		apihelper.WriteError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	currentUser, ok := requestctx.CurrentUserFromContext(r.Context())
	if !ok {
		apihelper.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	profile, err := h.service.UpdateProfile(r.Context(), currentUser.UserID, UpdateProfileInput{FullName: request.FullName})
	if err != nil {
		writeUserError(w, err)
		return
	}

	apihelper.WriteJSON(w, http.StatusOK, makeProfileResponse(profile))

}

func (h *Handler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := requestctx.CurrentUserFromContext(r.Context())
	if !ok {
		apihelper.WriteError(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	err := h.service.DeleteAccount(
		r.Context(),
		currentUser.UserID,
	)
	if err != nil {
		writeUserError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
