package listing

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Dimes2212/5th_semester_course_work_backend/internal/apihelper"
)

const maxListingPhotoRequestSize int64 = maxListingPhotoSize + (1 << 20)

func writeListingError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidListingID):
		apihelper.WriteError(
			w,
			http.StatusBadRequest,
			"invalid_listing_id",
		)

	case errors.Is(err, ErrListingNotFound):
		apihelper.WriteError(
			w,
			http.StatusNotFound,
			"listing_not_found",
		)

	case isValidationError(err):
		apihelper.WriteError(
			w,
			http.StatusBadRequest,
			"validation_error",
		)
	case errors.Is(err, ErrInvalidCursor):
		apihelper.WriteError(
			w,
			http.StatusBadRequest,
			"invalid_cursor",
		)
	case errors.Is(err, ErrInvalidPhotoRequest):
		apihelper.WriteError(
			w,
			http.StatusBadRequest,
			"invalid_photo_request",
		)

	case errors.Is(err, ErrEmptyPhoto):
		apihelper.WriteError(
			w,
			http.StatusBadRequest,
			"empty_photo",
		)

	case errors.Is(err, ErrPhotoTooLarge):
		apihelper.WriteError(
			w,
			http.StatusRequestEntityTooLarge,
			"photo_too_large",
		)

	case errors.Is(err, ErrInvalidPhotoType):
		apihelper.WriteError(
			w,
			http.StatusUnsupportedMediaType,
			"invalid_photo_type",
		)

	case errors.Is(err, ErrListingPhotoLimitReached):
		apihelper.WriteError(
			w,
			http.StatusConflict,
			"listing_photo_limit_reached",
		)
	case errors.Is(err, ErrInvalidListingPhotoID):
		apihelper.WriteError(
			w,
			http.StatusBadRequest,
			"invalid_listing_photo_id",
		)

	case errors.Is(err, ErrListingPhotoNotFound):
		apihelper.WriteError(
			w,
			http.StatusNotFound,
			"listing_photo_not_found",
		)
	default:
		apihelper.WriteError(
			w,
			http.StatusInternalServerError,
			"internal_error",
		)
	}
}

func isValidationError(err error) bool {
	return errors.Is(err, ErrCategoryRequired) ||
		errors.Is(err, ErrLocationRequired) ||
		errors.Is(err, ErrInvalidKind) ||
		errors.Is(err, ErrTitleRequired) ||
		errors.Is(err, ErrTitleTooLong) ||
		errors.Is(err, ErrDescriptionTooLong) ||
		errors.Is(err, ErrLocationDetailsTooLong) ||
		errors.Is(err, ErrEventAtRequired) ||
		errors.Is(err, ErrInvalidCategoryID) ||
		errors.Is(err, ErrInvalidLocationID) ||
		errors.Is(err, ErrSearchQueryTooLong)
}

func makeListingResponse(listing Listing) listingResponse {
	return listingResponse{
		ID:                 listing.ID,
		AuthorID:           listing.AuthorID,
		CategoryID:         listing.CategoryID,
		LocationID:         listing.LocationID,
		LocationDetails:    listing.LocationDetails,
		Kind:               listing.Kind,
		Title:              listing.Title,
		Description:        listing.Description,
		EventAt:            listing.EventAt,
		Status:             listing.Status,
		ResolvedWithUserID: listing.ResolvedWithUserID,
		ResolvedAt:         listing.ResolvedAt,
		CreatedAt:          listing.CreatedAt,
		UpdatedAt:          listing.UpdatedAt,
		Photos:             makeListingPhotoResponses(listing.Photos),
	}
}

func makeListingsFeedResponse(result GetListingsFeedResult) listingsFeedResponse {
	return listingsFeedResponse{
		Items:      makeListingsResponse(result.Listings),
		NextCursor: result.NextCursor,
		HasMore:    result.HasMore,
	}
}

func makeListingsResponse(
	listings []Listing,
) []listingResponse {
	items := make(
		[]listingResponse,
		0,
		len(listings),
	)

	for _, foundListing := range listings {
		items = append(
			items,
			makeListingResponse(foundListing),
		)
	}

	return items
}

func makeDeleteListingResponse(listing Listing) deleteListingResponse {
	return deleteListingResponse{
		ID:        listing.ID,
		DeletedAt: listing.DeletedAt,
	}
}

func parseGetListingsFeedInput(r *http.Request) (GetListingsFeedInput, error) {
	query := r.URL.Query()

	input := GetListingsFeedInput{
		Cursor: query.Get("cursor"),
		Query:  query.Get("q"),
	}

	kind := strings.TrimSpace(query.Get("kind"))
	if kind != "" {
		parsedKind := ListingKind(strings.ToLower(kind))
		input.Kind = &parsedKind
	}

	categoryID, err := parseOptionalInt16(query.Get("category_id"), ErrInvalidCategoryID)
	if err != nil {
		return GetListingsFeedInput{}, err
	}
	input.CategoryID = categoryID

	locationID, err := parseOptionalInt16(query.Get("location_id"), ErrInvalidLocationID)
	if err != nil {
		return GetListingsFeedInput{}, err
	}
	input.LocationID = locationID

	return input, nil
}

func parseOptionalInt16(
	value string,
	invalidValueError error,
) (*int16, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return nil, nil
	}

	parsedValue, err := strconv.ParseInt(value, 10, 16)
	if err != nil || parsedValue <= 0 {
		return nil, invalidValueError
	}

	result := int16(parsedValue)

	return &result, nil
}

func parseAddListingPhotoInput(w http.ResponseWriter, r *http.Request) (AddListingPhotoInput, error) {
	r.Body = http.MaxBytesReader(w, r.Body, maxListingPhotoRequestSize)

	if err := r.ParseMultipartForm(maxListingPhotoSize); err != nil {
		var maxBytesError *http.MaxBytesError

		if errors.As(err, &maxBytesError) {
			return AddListingPhotoInput{}, ErrPhotoTooLarge
		}

		return AddListingPhotoInput{}, ErrInvalidPhotoRequest
	}

	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	file, header, err := r.FormFile("photo")
	if err != nil {
		return AddListingPhotoInput{}, ErrInvalidPhotoRequest
	}
	defer file.Close()

	if header.Size <= 0 {
		return AddListingPhotoInput{}, ErrEmptyPhoto
	}

	if header.Size > maxListingPhotoSize {
		return AddListingPhotoInput{}, ErrPhotoTooLarge
	}

	data, err := io.ReadAll(io.LimitReader(file, maxListingPhotoSize+1))
	if err != nil {
		return AddListingPhotoInput{}, ErrInvalidPhotoRequest
	}

	if len(data) == 0 {
		return AddListingPhotoInput{}, ErrEmptyPhoto
	}

	if int64(len(data)) > maxListingPhotoSize {
		return AddListingPhotoInput{}, ErrPhotoTooLarge
	}

	contentType := http.DetectContentType(data)

	return AddListingPhotoInput{
		File:        bytes.NewReader(data),
		Size:        int64(len(data)),
		ContentType: contentType,
	}, nil
}

func makeAddListingPhotoResponse(result AddListingPhotoResult) addListingPhotoResponse {
	return addListingPhotoResponse{
		ID:       result.Photo.ID,
		URL:      result.URL,
		Position: result.Photo.Position,
	}
}

func makeListingPhotoResponses(
	photos []ListingPhoto,
) []listingPhotoResponse {
	response := make(
		[]listingPhotoResponse,
		0,
		len(photos),
	)

	for _, photo := range photos {
		response = append(
			response,
			listingPhotoResponse{
				ID:       photo.ID,
				URL:      photo.URL,
				Position: photo.Position,
			},
		)
	}

	return response
}
