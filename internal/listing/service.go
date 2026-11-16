package listing

import (
	"context"
	"errors"
	"fmt"
	"time"

	storage "github.com/Dimes2212/5th_semester_course_work_backend/internal/s3"
	"github.com/jackc/pgx/v5"
)

const listingPhotoURLTTL = 30 * time.Minute

type Service struct {
	repository *Repository
	storage    *storage.Client
}

func NewService(repository *Repository, storageClient *storage.Client) *Service {
	return &Service{
		repository: repository,
		storage:    storageClient,
	}
}

func (s *Service) CreateListing(ctx context.Context, authorID string, input CreateListingInput) (Listing, error) {

	if err := validateCreateListingInput(&input); err != nil {
		return Listing{}, err
	}

	createdListing, err := s.repository.CreateListing(ctx, authorID, input)
	if err != nil {
		return Listing{}, fmt.Errorf("create listing: %w", err)
	}

	return createdListing, nil

}

func (s *Service) GetListingByID(
	ctx context.Context,
	listingID string,
) (Listing, error) {
	if !isValidListingID(listingID) {
		return Listing{}, ErrInvalidListingID
	}

	foundListing, err := s.repository.FindListingByID(ctx, listingID)

	if errors.Is(err, pgx.ErrNoRows) {
		return Listing{}, ErrListingNotFound
	}

	if err != nil {
		return Listing{}, fmt.Errorf("find listing by id: %w", err)
	}

	listings := []Listing{foundListing}

	if err := s.attachPhotosToListings(ctx, listings); err != nil {
		return Listing{}, err
	}

	return listings[0], nil
}

func (s *Service) GetListingsFeed(ctx context.Context, input GetListingsFeedInput) (GetListingsFeedResult, error) {
	input, err := normalizeListingsFeedInput(input)
	if err != nil {
		return GetListingsFeedResult{}, err
	}

	cursor, err := decodeListingFeedCursor(input.Cursor)
	if err != nil {
		return GetListingsFeedResult{}, err
	}

	var searchQuery *string

	if input.Query != "" {
		searchQuery = &input.Query
	}

	listings, err := s.repository.FindActiveListingsPage(ctx, cursor,
		ListingFeedFilter{
			Query:      searchQuery,
			Kind:       input.Kind,
			CategoryID: input.CategoryID,
			LocationID: input.LocationID,
		},
		feedPageSize+1)
	if err != nil {
		return GetListingsFeedResult{}, fmt.Errorf("find listings page: %w", err)
	}

	hasMore := len(listings) > feedPageSize

	if hasMore {
		listings = listings[:feedPageSize]
	}

	var nextCursor *string

	if hasMore {
		encodedCursor, err := encodeListingFeedCursor(listings[len(listings)-1])
		if err != nil {
			return GetListingsFeedResult{}, err
		}

		nextCursor = &encodedCursor
	}

	if err := s.attachPhotosToListings(ctx, listings); err != nil {
		return GetListingsFeedResult{}, err
	}

	return GetListingsFeedResult{
		Listings:   listings,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil

}

func (s *Service) GetListingsByAuthorID(ctx context.Context, authorID string) ([]Listing, error) {
	listings, err := s.repository.FindListingsByAuthorID(ctx, authorID)
	if err != nil {
		return []Listing{}, fmt.Errorf("find author listings: %w", err)
	}

	if err := s.attachPhotosToListings(ctx, listings); err != nil {
		return []Listing{}, err
	}

	return listings, nil
}

func (s *Service) UpdateListing(ctx context.Context, listingID string, authorID string, input UpdateListingInput) (Listing, error) {
	if !isValidListingID(listingID) {
		return Listing{}, ErrInvalidListingID
	}

	if err := validateUpdateListingInput(&input); err != nil {
		return Listing{}, err
	}

	updatedListing, err := s.repository.UpdateListing(ctx, listingID, authorID, input)
	if errors.Is(err, pgx.ErrNoRows) {
		return Listing{}, ErrListingNotFound
	}

	if err != nil {
		return Listing{}, fmt.Errorf("update listing: %w", err)
	}
	listings := []Listing{updatedListing}

	if err := s.attachPhotosToListings(ctx, listings); err != nil {
		return Listing{}, err
	}

	return listings[0], nil

}

func (s *Service) ResolveListing(ctx context.Context, listingID string, authorID string) (Listing, error) {
	if !isValidListingID(listingID) {
		return Listing{}, ErrInvalidListingID
	}

	resolvedListing, err := s.repository.ResolveListing(
		ctx,
		listingID,
		authorID,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Listing{}, ErrListingNotFound
	}

	if err != nil {
		return Listing{}, fmt.Errorf(
			"resolve listing: %w",
			err,
		)
	}

	listings := []Listing{resolvedListing}

	if err := s.attachPhotosToListings(ctx, listings); err != nil {
		return Listing{}, err
	}

	return listings[0], nil

}

func (s *Service) DeleteListing(ctx context.Context, listingID string, authorID string) (Listing, error) {
	if !isValidListingID(listingID) {
		return Listing{}, ErrInvalidListingID
	}

	deletedListing, err := s.repository.DeleteListing(ctx, listingID, authorID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Listing{}, ErrListingNotFound
	}

	if err != nil {
		return Listing{}, fmt.Errorf(
			"delete listing: %w",
			err,
		)
	}
	return deletedListing, nil
}

func (s *Service) AddListingPhoto(ctx context.Context, listingID string, authorID string, input AddListingPhotoInput) (AddListingPhotoResult, error) {
	if !isValidListingID(listingID) {
		return AddListingPhotoResult{}, ErrInvalidListingID
	}

	if err := validateListingPhotoInput(input); err != nil {
		return AddListingPhotoResult{}, err
	}

	storageKey, err := generateListingPhotoStorageKey(listingID, input.ContentType)
	if err != nil {
		return AddListingPhotoResult{}, err
	}

	err = s.storage.Upload(ctx, storageKey, input.File, input.Size, input.ContentType)
	if err != nil {
		return AddListingPhotoResult{}, fmt.Errorf(
			"upload listing photo: %w",
			err,
		)
	}

	photoURL, err := s.storage.PresignedGetURL(ctx, storageKey, listingPhotoURLTTL)
	if err != nil {
		cleanupErr := s.cleanupListingPhoto(ctx, storageKey)
		resultErr := fmt.Errorf(
			"create listing photo URL: %w",
			err,
		)

		if cleanupErr != nil {
			resultErr = errors.Join(
				resultErr,
				fmt.Errorf(
					"cleanup uploaded listing photo: %w",
					cleanupErr,
				),
			)
		}

		return AddListingPhotoResult{}, resultErr
	}

	createdPhoto, err := s.repository.CreateListingPhoto(ctx, listingID, authorID, storageKey)
	if err != nil {
		cleanupErr := s.cleanupListingPhoto(ctx, storageKey)

		if errors.Is(err, pgx.ErrNoRows) {
			err = ErrListingNotFound
		} else if !errors.Is(err, ErrListingPhotoLimitReached) {
			err = fmt.Errorf("create listing photo: %w", err)
		}

		if cleanupErr != nil {
			err = errors.Join(err, fmt.Errorf("cleanup uploaded listing photo: %w", cleanupErr))
		}

		return AddListingPhotoResult{}, err
	}

	return AddListingPhotoResult{Photo: createdPhoto, URL: photoURL}, nil
}

func (s *Service) DeleteListingPhoto(ctx context.Context, listingID string, photoID string, authorID string) error {
	if !isValidListingID(listingID) {
		return ErrInvalidListingID
	}

	if !isValidListingPhotoID(photoID) {
		return ErrInvalidListingPhotoID
	}

	currentPhoto, err := s.repository.FindListingPhotoByIDAndOwner(ctx, listingID, photoID, authorID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrListingPhotoNotFound
	}
	if err != nil {
		return fmt.Errorf(
			"find listing photo for deletion: %w",
			err,
		)
	}

	storageCtx, cancel := context.WithTimeout(ctx, 10*time.Second)

	err = s.storage.DeleteObject(storageCtx, currentPhoto.StorageKey)

	cancel()
	if err != nil {
		return fmt.Errorf(
			"delete listing photo object: %w",
			err,
		)

	}

	err = s.repository.DeleteListingPhoto(ctx, listingID, photoID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrListingPhotoNotFound
	}
	if err != nil {
		return fmt.Errorf(
			"delete listing photo record: %w",
			err,
		)
	}
	return nil
}
