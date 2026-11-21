package admin

import (
	"context"

	"github.com/Dimes2212/5th_semester_course_work_backend/internal/listing"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindAllUsers(ctx context.Context) ([]UserDetails, error) {

	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			full_name,
			email,
			role,
			created_at,
			updated_at,
			deleted_at,
			blocked_at
		FROM users
		ORDER BY created_at DESC, id DESC
	`)

	if err != nil {
		return []UserDetails{}, err
	}
	defer rows.Close()

	users := make([]UserDetails, 0)

	for rows.Next() {
		var user UserDetails

		if err := rows.Scan(&user.ID,
			&user.FullName,
			&user.Email,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.DeletedAt,
			&user.BlockedAt,
		); err != nil {
			return []UserDetails{}, err
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil

}

func (r *Repository) SetUserBlocked(ctx context.Context, userID string, blocked bool) (UserDetails, error) {

	var updatedUser UserDetails

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return UserDetails{}, err
	}
	defer tx.Rollback(ctx)

	switch blocked {
	case true:
		err := tx.QueryRow(ctx, `
			UPDATE users
			SET blocked_at = now(), updated_at = now()
			WHERE id = $1 AND role = 'user' AND deleted_at IS NULL
			RETURNING
				id, full_name, email, role,created_at,updated_at, deleted_at, blocked_at;
		`, userID).Scan(&updatedUser.ID,
			&updatedUser.FullName,
			&updatedUser.Email,
			&updatedUser.Role,
			&updatedUser.CreatedAt,
			&updatedUser.UpdatedAt,
			&updatedUser.DeletedAt,
			&updatedUser.BlockedAt)
		if err != nil {
			return UserDetails{}, err
		}

		_, err = tx.Exec(ctx, `
			UPDATE refresh_sessions
			SET revoked_at = now()
			WHERE user_id = $1
			AND revoked_at IS NULL
		`, userID)
		if err != nil {
			return UserDetails{}, err
		}

	case false:
		err := tx.QueryRow(ctx, `
			UPDATE users
			SET blocked_at = NULL, updated_at = now()
			WHERE id = $1 AND role = 'user' AND deleted_at IS NULL
			RETURNING
				id, full_name, email, role,created_at,updated_at, deleted_at, blocked_at;
		`, userID).Scan(&updatedUser.ID,
			&updatedUser.FullName,
			&updatedUser.Email,
			&updatedUser.Role,
			&updatedUser.CreatedAt,
			&updatedUser.UpdatedAt,
			&updatedUser.DeletedAt,
			&updatedUser.BlockedAt)
		if err != nil {
			return UserDetails{}, err
		}

	}

	if err := tx.Commit(ctx); err != nil {
		return UserDetails{}, err
	}

	return updatedUser, nil
}

func (r *Repository) FindAllListings(ctx context.Context) ([]listing.Listing, error) {
	listings := make([]listing.Listing, 0)

	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			author_id,
			category_id,
			location_id,
			location_details,
			kind,
			title,
			description,
			event_at,
			status,
			resolved_with_user_id,
			resolved_at,
			created_at,
			updated_at,
			deleted_at
		FROM listings
		ORDER BY created_at DESC, id DESC
	`)

	if err != nil {
		return []listing.Listing{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var currentListing listing.Listing

		if err := rows.Scan(&currentListing.ID,
			&currentListing.AuthorID,
			&currentListing.CategoryID,
			&currentListing.LocationID,
			&currentListing.LocationDetails,
			&currentListing.Kind,
			&currentListing.Title,
			&currentListing.Description,
			&currentListing.EventAt,
			&currentListing.Status,
			&currentListing.ResolvedWithUserID,
			&currentListing.ResolvedAt,
			&currentListing.CreatedAt,
			&currentListing.UpdatedAt,
			&currentListing.DeletedAt,
		); err != nil {
			return []listing.Listing{}, err
		}

		listings = append(listings, currentListing)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return listings, nil
}

func (r *Repository) SoftDeleteListingByID(ctx context.Context, listingID string) (listing.Listing, error) {

	var deletedListing listing.Listing

	err := r.db.QueryRow(ctx, `
		UPDATE listings
		SET updated_at = now(), deleted_at = COALESCE(deleted_at, now())
		WHERE id = $1
		RETURNING
			id,
			author_id,
			category_id,
			location_id,
			location_details,
			kind,
			title,
			description,
			event_at,
			status,
			resolved_with_user_id,
			resolved_at,
			created_at,
			updated_at,
			deleted_at
	`, listingID).Scan(
		&deletedListing.ID,
		&deletedListing.AuthorID,
		&deletedListing.CategoryID,
		&deletedListing.LocationID,
		&deletedListing.LocationDetails,
		&deletedListing.Kind,
		&deletedListing.Title,
		&deletedListing.Description,
		&deletedListing.EventAt,
		&deletedListing.Status,
		&deletedListing.ResolvedWithUserID,
		&deletedListing.ResolvedAt,
		&deletedListing.CreatedAt,
		&deletedListing.UpdatedAt,
		&deletedListing.DeletedAt,
	)

	if err != nil {
		return listing.Listing{}, err
	}

	return deletedListing, nil
}

func (r *Repository) RestoreListingByID(ctx context.Context, listingID string) (listing.Listing, error) {
	var restoredListing listing.Listing

	err := r.db.QueryRow(ctx, `
		UPDATE listings
		SET deleted_at = NULL,
			updated_at = now()
		WHERE id = $1
		RETURNING
			id,
			author_id,
			category_id,
			location_id,
			location_details,
			kind,
			title,
			description,
			event_at,
			status,
			resolved_with_user_id,
			resolved_at,
			created_at,
			updated_at,
			deleted_at
	`, listingID).Scan(
		&restoredListing.ID,
		&restoredListing.AuthorID,
		&restoredListing.CategoryID,
		&restoredListing.LocationID,
		&restoredListing.LocationDetails,
		&restoredListing.Kind,
		&restoredListing.Title,
		&restoredListing.Description,
		&restoredListing.EventAt,
		&restoredListing.Status,
		&restoredListing.ResolvedWithUserID,
		&restoredListing.ResolvedAt,
		&restoredListing.CreatedAt,
		&restoredListing.UpdatedAt,
		&restoredListing.DeletedAt,
	)

	if err != nil {
		return listing.Listing{}, err
	}

	return restoredListing, nil
}
