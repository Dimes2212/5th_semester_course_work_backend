package admin

import (
	"errors"
	"net/http"

	"github.com/Dimes2212/5th_semester_course_work_backend/internal/apihelper"
	"github.com/Dimes2212/5th_semester_course_work_backend/internal/listing"
)

func parseBlockedQuery(r *http.Request) (bool, error) {
	switch r.URL.Query().Get("blocked") {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, ErrInvalidBlockedValue
	}
}

func writeAdminError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidUserID),
		errors.Is(err, ErrInvalidBlockedValue):
		apihelper.WriteError(w, http.StatusBadRequest, "invalid_request")

	case errors.Is(err, ErrUserNotFound):
		apihelper.WriteError(w, http.StatusNotFound, "user_not_found")
	case errors.Is(err, ErrInvalidListingID):
		apihelper.WriteError(w, http.StatusBadRequest, "invalid_listing_id")

	case errors.Is(err, ErrListingNotFound):
		apihelper.WriteError(w, http.StatusNotFound, "listing_not_found")

	default:
		apihelper.WriteError(w, http.StatusInternalServerError, "internal_error")
	}
}

func makeListingResponse(currentListing listing.Listing) listingResponse {
	return listingResponse{
		ID:                 currentListing.ID,
		AuthorID:           currentListing.AuthorID,
		CategoryID:         currentListing.CategoryID,
		LocationID:         currentListing.LocationID,
		LocationDetails:    currentListing.LocationDetails,
		Kind:               currentListing.Kind,
		Title:              currentListing.Title,
		Description:        currentListing.Description,
		EventAt:            currentListing.EventAt,
		Status:             currentListing.Status,
		ResolvedWithUserID: currentListing.ResolvedWithUserID,
		ResolvedAt:         currentListing.ResolvedAt,
		CreatedAt:          currentListing.CreatedAt,
		UpdatedAt:          currentListing.UpdatedAt,
		DeletedAt:          currentListing.DeletedAt,
		Photos:             makeListingPhotoResponses(currentListing.Photos),
	}
}

func makeListingResponses(listings []listing.Listing) []listingResponse {
	responses := make([]listingResponse, 0, len(listings))

	for _, currentListing := range listings {
		responses = append(responses, makeListingResponse(currentListing))
	}

	return responses
}

func makeUserResponse(user UserDetails) userResponse {
	return userResponse{
		ID:        user.ID,
		FullName:  user.FullName,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		DeletedAt: user.DeletedAt,
		BlockedAt: user.BlockedAt,
	}
}

func makeUserResponses(users []UserDetails) []userResponse {
	responses := make([]userResponse, 0, len(users))

	for _, user := range users {
		responses = append(responses, makeUserResponse(user))
	}

	return responses
}

func makeListingPhotoResponses(
	photos []listing.ListingPhoto,
) []listingPhotoResponse {
	responses := make([]listingPhotoResponse, 0, len(photos))

	for _, photo := range photos {
		responses = append(responses, listingPhotoResponse{
			ID:       photo.ID,
			URL:      photo.URL,
			Position: photo.Position,
		})
	}

	return responses
}
