package lookup

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindAllCategories(ctx context.Context) ([]Category, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, slug, name
		FROM categories
		WHERE is_active = true
		ORDER BY position, id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]Category, 0)

	for rows.Next() {
		var category Category

		if err := rows.Scan(&category.ID, &category.Slug, &category.Name); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *Repository) FindAllLocations(
	ctx context.Context,
) ([]Location, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			name,
			address
		FROM locations
		WHERE is_active = true
		ORDER BY position, id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	locations := make([]Location, 0)

	for rows.Next() {
		var location Location

		if err := rows.Scan(
			&location.ID,
			&location.Name,
			&location.Address,
		); err != nil {
			return nil, err
		}

		locations = append(locations, location)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return locations, nil
}
