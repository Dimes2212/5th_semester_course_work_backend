package listing

import (
	"net/http"

	"github.com/Dimes2212/5th_semester_course_work_backend/internal/apihelper"
	"github.com/Dimes2212/5th_semester_course_work_backend/internal/requestctx"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) CreateListing(w http.ResponseWriter, r *http.Request) {
	var request createListingRequest

	if err := apihelper.DecodeJSON(w, r, &request); err != nil {
		apihelper.WriteError(
			w,
			http.StatusBadRequest,
			"invalid_request",
		)
		return
	}

	currentUser, ok := requestctx.CurrentUserFromContext(r.Context())
	if !ok {
		apihelper.WriteError(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	createdListing, err := h.service.CreateListing(
		r.Context(),
		currentUser.UserID,
		CreateListingInput{
			CategoryID:      request.CategoryID,
			LocationID:      request.LocationID,
			LocationDetails: request.LocationDetails,
			Kind:            request.Kind,
			Title:           request.Title,
			Description:     request.Description,
			EventAt:         request.EventAt,
		},
	)
	if err != nil {
		writeListingError(w, err)
		return
	}

	apihelper.WriteJSON(
		w,
		http.StatusCreated,
		makeListingResponse(createdListing),
	)
}

func (h *Handler) GetListingByID(w http.ResponseWriter, r *http.Request) {
	listingID := chi.URLParam(r, "listingID")

	foundListing, err := h.service.GetListingByID(r.Context(), listingID)

	if err != nil {
		writeListingError(w, err)
		return
	}

	apihelper.WriteJSON(w, http.StatusOK, makeListingResponse(foundListing))
}

func (h *Handler) GetListingsFeed(w http.ResponseWriter, r *http.Request) {
	input, err := parseGetListingsFeedInput(r)
	if err != nil {
		writeListingError(w, err)
		return
	}

	result, err := h.service.GetListingsFeed(
		r.Context(),
		input,
	)
	if err != nil {
		writeListingError(w, err)
		return
	}

	apihelper.WriteJSON(
		w,
		http.StatusOK,
		makeListingsFeedResponse(result),
	)
}

func (h *Handler) GetMyListings(w http.ResponseWriter, r *http.Request) {

	currentUser, ok := requestctx.CurrentUserFromContext(r.Context())
	if !ok {
		apihelper.WriteError(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	result, err := h.service.GetListingsByAuthorID(r.Context(), currentUser.UserID)
	if err != nil {
		writeListingError(w, err)
		return
	}

	apihelper.WriteJSON(w, http.StatusOK, makeListingsResponse(result))
}

func (h *Handler) UpdateListing(w http.ResponseWriter, r *http.Request) {
	author, ok := requestctx.CurrentUserFromContext(r.Context())
	if !ok {
		apihelper.WriteError(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}
	var request updateListingRequest
	if err := apihelper.DecodeJSON(w, r, &request); err != nil {
		apihelper.WriteError(
			w,
			http.StatusBadRequest,
			"invalid_request",
		)
		return
	}

	listingID := chi.URLParam(r, "listingID")

	result, err := h.service.UpdateListing(r.Context(), listingID, author.UserID, UpdateListingInput(request))
	if err != nil {
		writeListingError(w, err)
		return
	}

	apihelper.WriteJSON(
		w,
		http.StatusOK,
		makeListingResponse(result),
	)

}

func (h *Handler) ResolveListing(w http.ResponseWriter, r *http.Request) {
	listingID := chi.URLParam(r, "listingID")

	currentUser, ok := requestctx.CurrentUserFromContext(r.Context())
	if !ok {
		apihelper.WriteError(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	resolvedListing, err := h.service.ResolveListing(r.Context(), listingID, currentUser.UserID)
	if err != nil {
		writeListingError(w, err)
		return
	}

	apihelper.WriteJSON(
		w,
		http.StatusOK,
		makeListingResponse(resolvedListing),
	)
}

func (h *Handler) DeleteListing(w http.ResponseWriter, r *http.Request) {
	listingID := chi.URLParam(r, "listingID")

	currentUser, ok := requestctx.CurrentUserFromContext(r.Context())
	if !ok {
		apihelper.WriteError(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	deletedListing, err := h.service.DeleteListing(r.Context(), listingID, currentUser.UserID)
	if err != nil {
		writeListingError(w, err)
		return
	}

	apihelper.WriteJSON(
		w,
		http.StatusOK,
		makeDeleteListingResponse(deletedListing),
	)
}

func (h *Handler) AddListingPhoto(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := requestctx.CurrentUserFromContext(r.Context())
	if !ok {
		apihelper.WriteError(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	listingID := chi.URLParam(r, "listingID")

	input, err := parseAddListingPhotoInput(w, r)
	if err != nil {
		writeListingError(w, err)
		return
	}

	result, err := h.service.AddListingPhoto(r.Context(), listingID, currentUser.UserID, input)
	if err != nil {
		writeListingError(w, err)
		return
	}

	apihelper.WriteJSON(w, http.StatusCreated, makeAddListingPhotoResponse(result))
}

func (h *Handler) DeleteListingPhoto(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := requestctx.CurrentUserFromContext(r.Context())
	if !ok {
		apihelper.WriteError(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	listingID := chi.URLParam(r, "listingID")
	photoID := chi.URLParam(r, "photoID")

	err := h.service.DeleteListingPhoto(r.Context(), listingID, photoID, currentUser.UserID)
	if err != nil {
		writeListingError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}
