package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) GetByID(ctx context.Context, userID string) (Profile, error) {
	foundUser, err := s.repository.FindByID(ctx, userID)

	if errors.Is(err, pgx.ErrNoRows) {
		return Profile{}, ErrUserNotFound
	}

	if err != nil {
		return Profile{}, fmt.Errorf(
			"find user by id: %w",
			err,
		)
	}

	return makeProfile(foundUser), nil
}

func (s *Service) UpdateProfile(
	ctx context.Context,
	userID string,
	input UpdateProfileInput,
) (Profile, error) {
	fullName := strings.TrimSpace(input.FullName)

	if fullName == "" {
		return Profile{}, ErrFullNameRequired
	}
	if utf8.RuneCountInString(fullName) > 100 {
		return Profile{}, ErrFullNameTooLong
	}

	updatedUser, err := s.repository.UpdateFullName(ctx, userID, fullName)
	if errors.Is(err, pgx.ErrNoRows) {
		return Profile{}, ErrUserNotFound
	}

	if err != nil {
		return Profile{}, fmt.Errorf(
			"update user full name: %w",
			err,
		)
	}

	return makeProfile(updatedUser), nil
}

func (s *Service) DeleteAccount(ctx context.Context, userID string) error {
	err := s.repository.DeleteByID(ctx, userID)

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrUserNotFound
	}

	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	return nil
}
