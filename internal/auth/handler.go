package auth

import (
	"net/http"
	"time"

	"github.com/Dimes2212/5th_semester_course_work_backend/internal/user"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
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

type errorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var request registerRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request")
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

	writeJSON(w, http.StatusCreated, makeAuthResponse(result))
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request")
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

	writeJSON(w, http.StatusOK, makeAuthResponse(result))
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var request refreshRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	result, err := h.service.Refresh(r.Context(), request.RefreshToken)
	if err != nil {
		writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, makeAuthResponse(result))
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var request logoutRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	if err := h.service.Logout(r.Context(), request.RefreshToken); err != nil {
		writeAuthError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
