package admin

import (
	"context"
	"errors"
	"fmt"

	"github.com/Dimes2212/5th_semester_course_work_backend/internal/listing"
	storage "github.com/Dimes2212/5th_semester_course_work_backend/internal/s3"
	"github.com/jackc/pgx/v5"
)

type Service struct {
	repository        *Repository
	listingRepository *listing.Repository
	storage           *storage.Client
}

func NewService(repository *Repository, listingRepository *listing.Repository, storageClient *storage.Client) *Service {
	return &Service{
		repository:        repository,
		listingRepository: listingRepository,
		storage:           storageClient,
	}
}

func (s *Service) GetAllUsers(ctx context.Context) ([]UserDetails, error) {
	users, err := s.repository.FindAllUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("find all users: %w", err)
	}

	return users, nil
}

func (s *Service) SetUserBlocked(ctx context.Context, userID string, blocked bool) (UserDetails, error) {
	if !isValidUUID(userID) {
		return UserDetails{}, ErrInvalidUserID
	}

	updatedUser, err := s.repository.SetUserBlocked(ctx, userID, blocked)
	if errors.Is(err, pgx.ErrNoRows) {
		return UserDetails{}, ErrUserNotFound
	}

	if err != nil {
		return UserDetails{}, fmt.Errorf(
			"set user blocked status: %w",
			err,
		)
	}

	return updatedUser, nil
}

func (s *Service) GetAllListings(
	ctx context.Context,
) ([]listing.Listing, error) {
	listings, err := s.repository.FindAllListings(ctx)
	if err != nil {
		return nil, fmt.Errorf("find all listings: %w", err)
	}

	if err := s.attachPhotosToListings(ctx, listings); err != nil {
		return nil, err
	}

	return listings, nil
}

func (s *Service) SoftDeleteListingByID(
	ctx context.Context,
	listingID string,
) (listing.Listing, error) {
	if !isValidUUID(listingID) {
		return listing.Listing{}, ErrInvalidListingID
	}

	deletedListing, err := s.repository.SoftDeleteListingByID(
		ctx,
		listingID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return listing.Listing{}, ErrListingNotFound
	}

	if err != nil {
		return listing.Listing{}, fmt.Errorf(
			"soft delete listing by id: %w",
			err,
		)
	}

	listings := []listing.Listing{deletedListing}

	if err := s.attachPhotosToListings(ctx, listings); err != nil {
		return listing.Listing{}, err
	}

	return listings[0], nil

}

func (s *Service) RestoreListingByID(
	ctx context.Context,
	listingID string,
) (listing.Listing, error) {
	if !isValidUUID(listingID) {
		return listing.Listing{}, ErrInvalidListingID
	}

	restoredListing, err := s.repository.RestoreListingByID(
		ctx,
		listingID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return listing.Listing{}, ErrListingNotFound
	}

	if err != nil {
		return listing.Listing{}, fmt.Errorf(
			"restore listing by id: %w",
			err,
		)
	}

	listings := []listing.Listing{restoredListing}

	if err := s.attachPhotosToListings(ctx, listings); err != nil {
		return listing.Listing{}, err
	}

	return listings[0], nil

}
