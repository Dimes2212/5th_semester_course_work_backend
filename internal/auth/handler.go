package auth

import (
	"net/http"

	"github.com/Dimes2212/5th_semester_course_work_backend/internal/apihelper"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var request registerRequest
	if err := apihelper.DecodeJSON(w, r, &request); err != nil {
		apihelper.WriteError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	result, err := h.service.Register(r.Context(), RegisterInput{
		FullName:             request.FullName,
		Email:                request.Email,
		Password:             request.Password,
		PasswordConfirmation: request.PasswordConfirmation,
		DeviceName:           request.DeviceName,
	})
	if err != nil {
		writeAuthError(w, err)
		return
	}

	apihelper.WriteJSON(w, http.StatusCreated, makeAuthResponse(result))
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if err := apihelper.DecodeJSON(w, r, &request); err != nil {
		apihelper.WriteError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	result, err := h.service.Login(r.Context(), LoginInput{
		Email:      request.Email,
		Password:   request.Password,
		DeviceName: request.DeviceName,
	})
	if err != nil {
		writeAuthError(w, err)
		return
	}

	apihelper.WriteJSON(w, http.StatusOK, makeAuthResponse(result))
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var request refreshRequest
	if err := apihelper.DecodeJSON(w, r, &request); err != nil {
		apihelper.WriteError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	result, err := h.service.Refresh(r.Context(), request.RefreshToken)
	if err != nil {
		writeAuthError(w, err)
		return
	}

	apihelper.WriteJSON(w, http.StatusOK, makeAuthResponse(result))
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var request logoutRequest
	if err := apihelper.DecodeJSON(w, r, &request); err != nil {
		apihelper.WriteError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	if err := h.service.Logout(r.Context(), request.RefreshToken); err != nil {
		writeAuthError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
