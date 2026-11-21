package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/Dimes2212/5th_semester_course_work_backend/internal/listing"
	"github.com/jackc/pgx/v5/pgtype"
)

const listingPhotoURLTTL = 30 * time.Minute

func isValidUUID(value string) bool {
	var id pgtype.UUID

	if err := id.Scan(value); err != nil {
		return false
	}

	return id.Valid
}

func (s *Service) attachPhotosToListings(ctx context.Context, listings []listing.Listing) error {
	if len(listings) == 0 {
		return nil
	}

	listingIDs := make([]string, 0, len(listings))
	listingIndexes := make(map[string]int, len(listings))

	for index := range listings {
		listings[index].Photos = make([]listing.ListingPhoto, 0)
		listingIDs = append(listingIDs, listings[index].ID)
		listingIndexes[listings[index].ID] = index
	}

	photos, err := s.listingRepository.FindListingPhotosByListingIDs(
		ctx,
		listingIDs,
	)
	if err != nil {
		return fmt.Errorf("find listing photos: %w", err)
	}

	for _, photo := range photos {
		photoURL, err := s.storage.PresignedGetURL(
			ctx,
			photo.StorageKey,
			listingPhotoURLTTL,
		)
		if err != nil {
			return fmt.Errorf("create listing photo URL: %w", err)
		}

		photo.URL = photoURL

		index, ok := listingIndexes[photo.ListingID]
		if !ok {
			continue
		}

		listings[index].Photos = append(listings[index].Photos, photo)
	}

	return nil
}
