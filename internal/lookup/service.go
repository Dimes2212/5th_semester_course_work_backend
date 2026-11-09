package lookup

import (
	"context"
	"fmt"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) GetCategories(
	ctx context.Context,
) ([]Category, error) {
	categories, err := s.repository.FindAllCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("find all categories: %w", err)
	}

	return categories, nil
}

func (s *Service) GetLocations(
	ctx context.Context,
) ([]Location, error) {
	locations, err := s.repository.FindAllLocations(ctx)
	if err != nil {
		return nil, fmt.Errorf("find all locations: %w", err)
	}

	return locations, nil
}
