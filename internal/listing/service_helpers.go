package listing

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5/pgtype"
)

const feedPageSize = 4
const maxListingPhotoSize int64 = 5 << 20 // 5 МБ

func validateCreateListingInput(input *CreateListingInput) error {
	if input.CategoryID <= 0 {
		return ErrCategoryRequired
	}

	if input.LocationID <= 0 {
		return ErrLocationRequired
	}

	if input.Kind != ListingKindLost && input.Kind != ListingKindFound {
		return ErrInvalidKind
	}

	input.Title = strings.TrimSpace(input.Title)

	if input.Title == "" {
		return ErrTitleRequired
	}

	if utf8.RuneCountInString(input.Title) > 120 {
		return ErrTitleTooLong
	}

	input.Description = strings.TrimSpace(input.Description)

	if utf8.RuneCountInString(input.Description) > 2000 {
		return ErrDescriptionTooLong
	}

	if input.LocationDetails != nil {
		locationDetails := strings.TrimSpace(*input.LocationDetails)

		if locationDetails == "" {
			input.LocationDetails = nil
		} else {
			if utf8.RuneCountInString(locationDetails) > 150 {
				return ErrLocationDetailsTooLong
			}

			input.LocationDetails = &locationDetails
		}
	}

	if input.EventAt.IsZero() {
		return ErrEventAtRequired
	}

	return nil
}

func isValidListingID(listingID string) bool {
	var id pgtype.UUID

	if err := id.Scan(listingID); err != nil {
		return false
	}

	return id.Valid
}

func decodeListingFeedCursor(encodedCursor string) (*ListingFeedCursor, error) {
	if encodedCursor == "" {
		return nil, nil
	}
	cursorJSON, err := base64.RawURLEncoding.DecodeString(encodedCursor)
	if err != nil {
		return nil, ErrInvalidCursor
	}

	var cursor ListingFeedCursor

	if err := json.Unmarshal(cursorJSON, &cursor); err != nil {
		return nil, ErrInvalidCursor
	}
	if cursor.CreatedAt.IsZero() || !isValidListingID(cursor.ID) {
		return nil, ErrInvalidCursor
	}

	return &cursor, nil
}

func encodeListingFeedCursor(listing Listing) (string, error) {
	cursorJSON, err := json.Marshal(ListingFeedCursor{
		CreatedAt: listing.CreatedAt,
		ID:        listing.ID,
	})
	if err != nil {
		return "", fmt.Errorf("marshal list cursor: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(cursorJSON), nil
}

func validateUpdateListingInput(
	input *UpdateListingInput,
) error {
	normalizedInput := CreateListingInput(*input)

	if err := validateCreateListingInput(
		&normalizedInput,
	); err != nil {
		return err
	}

	*input = UpdateListingInput(normalizedInput)

	return nil
}

func normalizeListingsFeedInput(input GetListingsFeedInput) (GetListingsFeedInput, error) {
	input.Query = strings.TrimSpace(input.Query)

	if utf8.RuneCountInString(input.Query) > 100 {
		return GetListingsFeedInput{}, ErrSearchQueryTooLong
	}

	if input.Kind != nil && *input.Kind != ListingKindLost && *input.Kind != ListingKindFound {
		return GetListingsFeedInput{}, ErrInvalidKind
	}

	if input.CategoryID != nil && *input.CategoryID <= 0 {
		return GetListingsFeedInput{}, ErrInvalidCategoryID
	}

	if input.LocationID != nil && *input.LocationID <= 0 {
		return GetListingsFeedInput{}, ErrInvalidLocationID
	}

	return input, nil
}

func validateListingPhotoInput(input AddListingPhotoInput) error {
	if input.File == nil || input.Size <= 0 {
		return ErrEmptyPhoto
	}
	if input.Size > maxListingPhotoSize {
		return ErrPhotoTooLarge
	}

	switch input.ContentType {
	case "image/jpeg", "image/png", "image/webp":
		return nil
	default:
		return ErrInvalidPhotoType
	}
}

func listingPhotoExtension(contentType string) (string, error) {
	switch contentType {
	case "image/jpeg":
		return ".jpg", nil
	case "image/png":
		return ".png", nil
	case "image/webp":
		return ".webp", nil
	default:
		return "", ErrInvalidPhotoType
	}
}

func generateListingPhotoStorageKey(listingID string, contentType string) (string, error) {
	extension, err := listingPhotoExtension(contentType)
	if err != nil {
		return "", err
	}

	randomBytes := make([]byte, 16)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate listing photo storage key: %w", err)
	}

	randomName := hex.EncodeToString(randomBytes)

	return fmt.Sprintf(
		"listings/%s/%s%s",
		listingID,
		randomName,
		extension,
	), nil
}

func (s *Service) cleanupListingPhoto(
	ctx context.Context,
	storageKey string,
) error {
	cleanupCtx, cancel := context.WithTimeout(
		context.WithoutCancel(ctx),
		10*time.Second,
	)
	defer cancel()

	return s.storage.DeleteObject(
		cleanupCtx,
		storageKey,
	)
}

func (s *Service) attachPhotosToListings(ctx context.Context, listings []Listing) error {
	if len(listings) == 0 {
		return nil
	}

	listingsIDs := make([]string, 0, len(listings))
	listingIndexes := make(map[string]int, len(listings))

	for index := range listings {
		listings[index].Photos = make([]ListingPhoto, 0)

		listingsIDs = append(listingsIDs, listings[index].ID)

		listingIndexes[listings[index].ID] = index
	}

	photos, err := s.repository.FindListingPhotosByListingIDs(ctx, listingsIDs)
	if err != nil {
		return fmt.Errorf("find listing photos: %w", err)
	}

	for _, photo := range photos {
		photoURL, err := s.storage.PresignedGetURL(ctx, photo.StorageKey, listingPhotoURLTTL)
		if err != nil {
			return fmt.Errorf("create listing photo URL: %w", err)
		}

		photo.URL = photoURL

		listingIndex, ok := listingIndexes[photo.ListingID]
		if !ok {
			continue
		}

		listings[listingIndex].Photos = append(listings[listingIndex].Photos, photo)

	}

	return nil

}

func isValidListingPhotoID(photoID string) bool {
	var id pgtype.UUID
	if err := id.Scan(photoID); err != nil {
		return false
	}

	return true
}
