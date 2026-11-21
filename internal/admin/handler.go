package admin

import (
	"net/http"

	"github.com/Dimes2212/5th_semester_course_work_backend/internal/apihelper"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.GetAllUsers(r.Context())
	if err != nil {
		apihelper.WriteError(
			w,
			http.StatusInternalServerError,
			"internal_error",
		)
		return
	}

	apihelper.WriteJSON(
		w,
		http.StatusOK,
		makeUserResponses(users),
	)
}

func (h *Handler) SetUserBlocked(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	blocked, err := parseBlockedQuery(r)
	if err != nil {
		writeAdminError(w, err)
		return
	}

	updatedUser, err := h.service.SetUserBlocked(
		r.Context(),
		userID,
		blocked,
	)
	if err != nil {
		writeAdminError(w, err)
		return
	}

	apihelper.WriteJSON(
		w,
		http.StatusOK,
		makeUserResponse(updatedUser),
	)
}

func (h *Handler) GetAllListings(w http.ResponseWriter, r *http.Request) {
	listings, err := h.service.GetAllListings(r.Context())
	if err != nil {
		writeAdminError(w, err)
		return
	}

	apihelper.WriteJSON(
		w,
		http.StatusOK,
		makeListingResponses(listings),
	)
}

func (h *Handler) SoftDeleteListingByID(w http.ResponseWriter, r *http.Request) {
	listingID := chi.URLParam(r, "listingID")

	deletedListing, err := h.service.SoftDeleteListingByID(
		r.Context(),
		listingID,
	)
	if err != nil {
		writeAdminError(w, err)
		return
	}

	apihelper.WriteJSON(
		w,
		http.StatusOK,
		makeListingResponse(deletedListing),
	)
}

func (h *Handler) RestoreListingByID(w http.ResponseWriter, r *http.Request) {
	listingID := chi.URLParam(r, "listingID")

	restoredListing, err := h.service.RestoreListingByID(
		r.Context(),
		listingID,
	)
	if err != nil {
		writeAdminError(w, err)
		return
	}

	apihelper.WriteJSON(
		w,
		http.StatusOK,
		makeListingResponse(restoredListing),
	)
}
