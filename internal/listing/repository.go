package listing

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) CreateListing(
	ctx context.Context,
	authorID string,
	input CreateListingInput,
) (Listing, error) {
	var createdListing Listing

	err := r.db.QueryRow(ctx, `
		INSERT INTO listings (
			author_id,
			category_id,
			location_id,
			location_details,
			kind,
			title,
			description,
			event_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
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
	`,
		authorID,
		input.CategoryID,
		input.LocationID,
		input.LocationDetails,
		input.Kind,
		input.Title,
		input.Description,
		input.EventAt,
	).Scan(
		&createdListing.ID,
		&createdListing.AuthorID,
		&createdListing.CategoryID,
		&createdListing.LocationID,
		&createdListing.LocationDetails,
		&createdListing.Kind,
		&createdListing.Title,
		&createdListing.Description,
		&createdListing.EventAt,
		&createdListing.Status,
		&createdListing.ResolvedWithUserID,
		&createdListing.ResolvedAt,
		&createdListing.CreatedAt,
		&createdListing.UpdatedAt,
		&createdListing.DeletedAt,
	)
	if err != nil {
		return Listing{}, err
	}

	return createdListing, nil
}

func (r *Repository) FindListingByID(ctx context.Context, listingID string) (Listing, error) {
	var foundListing Listing

	err := r.db.QueryRow(ctx, `
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
		WHERE id = $1
		  AND deleted_at IS NULL
	`, listingID).Scan(
		&foundListing.ID,
		&foundListing.AuthorID,
		&foundListing.CategoryID,
		&foundListing.LocationID,
		&foundListing.LocationDetails,
		&foundListing.Kind,
		&foundListing.Title,
		&foundListing.Description,
		&foundListing.EventAt,
		&foundListing.Status,
		&foundListing.ResolvedWithUserID,
		&foundListing.ResolvedAt,
		&foundListing.CreatedAt,
		&foundListing.UpdatedAt,
		&foundListing.DeletedAt,
	)

	return foundListing, err
}

func (r *Repository) FindActiveListingsPage(
	ctx context.Context,
	cursor *ListingFeedCursor,
	filter ListingFeedFilter,
	limit int,
) ([]Listing, error) {

	var cursorCreatedAt any
	var cursorID any

	if cursor != nil {
		cursorCreatedAt = cursor.CreatedAt
		cursorID = cursor.ID
	}

	rows, err := r.db.Query(ctx, `
		WITH page AS (
			SELECT id,
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
			WHERE status = 'active'
				AND deleted_at IS NULL
				AND (
					$1::timestamptz IS NULL
					OR (created_at, id) < ($1::timestamptz, $2::uuid)
				)
				AND ($3::varchar IS NULL OR kind = $3)
				AND ($4::smallint IS NULL OR category_id = $4)
				AND ($5::smallint IS NULL OR location_id = $5)
				AND ($6::text IS NULL OR title ILIKE '%' || $6 || '%' OR description ILIKE '%' || $6 || '%')
			ORDER BY created_at DESC, id DESC
			LIMIT $7
		)
		SELECT *
		FROM page
		ORDER BY created_at DESC, id DESC
	`, cursorCreatedAt, cursorID, filter.Kind, filter.CategoryID, filter.LocationID, filter.Query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	listings := make([]Listing, 0)

	for rows.Next() {
		var listing Listing

		if err := rows.Scan(
			&listing.ID,
			&listing.AuthorID,
			&listing.CategoryID,
			&listing.LocationID,
			&listing.LocationDetails,
			&listing.Kind,
			&listing.Title,
			&listing.Description,
			&listing.EventAt,
			&listing.Status,
			&listing.ResolvedWithUserID,
			&listing.ResolvedAt,
			&listing.CreatedAt,
			&listing.UpdatedAt,
			&listing.DeletedAt,
		); err != nil {
			return nil, err
		}
		listings = append(listings, listing)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return listings, nil
}

func (r *Repository) FindListingsByAuthorID(ctx context.Context, authorID string) ([]Listing, error) {
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
		WHERE author_id = $1
			AND deleted_at IS NULL
		ORDER BY created_at DESC, id DESC
	`, authorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	listings := make([]Listing, 0)

	for rows.Next() {
		var listing Listing

		if err := rows.Scan(
			&listing.ID,
			&listing.AuthorID,
			&listing.CategoryID,
			&listing.LocationID,
			&listing.LocationDetails,
			&listing.Kind,
			&listing.Title,
			&listing.Description,
			&listing.EventAt,
			&listing.Status,
			&listing.ResolvedWithUserID,
			&listing.ResolvedAt,
			&listing.CreatedAt,
			&listing.UpdatedAt,
			&listing.DeletedAt,
		); err != nil {
			return nil, err
		}
		listings = append(listings, listing)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return listings, nil

}

func (r *Repository) UpdateListing(ctx context.Context, listingID string, authorID string, input UpdateListingInput) (Listing, error) {
	var updatedListing Listing

	err := r.db.QueryRow(ctx, `
	UPDATE listings
	SET
		category_id = $3,
		location_id = $4,
		location_details = $5,
		kind = $6,
		title = $7,
		description = $8,
		event_at = $9,
		updated_at = now()
	WHERE id = $1
	  AND author_id = $2
	  AND status = 'active'
	  AND deleted_at IS NULL
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
`,
		listingID,
		authorID,
		input.CategoryID,
		input.LocationID,
		input.LocationDetails,
		input.Kind,
		input.Title,
		input.Description,
		input.EventAt,
	).Scan(
		&updatedListing.ID,
		&updatedListing.AuthorID,
		&updatedListing.CategoryID,
		&updatedListing.LocationID,
		&updatedListing.LocationDetails,
		&updatedListing.Kind,
		&updatedListing.Title,
		&updatedListing.Description,
		&updatedListing.EventAt,
		&updatedListing.Status,
		&updatedListing.ResolvedWithUserID,
		&updatedListing.ResolvedAt,
		&updatedListing.CreatedAt,
		&updatedListing.UpdatedAt,
		&updatedListing.DeletedAt,
	)
	return updatedListing, err
}

func (r *Repository) ResolveListing(ctx context.Context, listingID string, authorID string) (Listing, error) {
	var resolvedListing Listing

	err := r.db.QueryRow(ctx, `
		UPDATE listings
		SET
			status = 'resolved',
			resolved_at = now(),
			updated_at = now()
		WHERE id = $1
		AND author_id = $2
		AND status = 'active'
		AND deleted_at IS NULL
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
	`, listingID, authorID).Scan(
		&resolvedListing.ID,
		&resolvedListing.AuthorID,
		&resolvedListing.CategoryID,
		&resolvedListing.LocationID,
		&resolvedListing.LocationDetails,
		&resolvedListing.Kind,
		&resolvedListing.Title,
		&resolvedListing.Description,
		&resolvedListing.EventAt,
		&resolvedListing.Status,
		&resolvedListing.ResolvedWithUserID,
		&resolvedListing.ResolvedAt,
		&resolvedListing.CreatedAt,
		&resolvedListing.UpdatedAt,
		&resolvedListing.DeletedAt,
	)
	return resolvedListing, err
}

func (r *Repository) DeleteListing(ctx context.Context, listingID string, authorID string) (Listing, error) {
	var deletedListing Listing

	err := r.db.QueryRow(ctx, `
		UPDATE listings
		SET
			deleted_at = now(),
			updated_at = now()
		WHERE id = $1
			AND author_id = $2
			AND deleted_at IS NULL
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
	`, listingID, authorID).Scan(
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
	return deletedListing, err
}

func (r *Repository) CreateListingPhoto(ctx context.Context, listingID string, authorID string, storageKey string) (ListingPhoto, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return ListingPhoto{}, err
	}
	defer tx.Rollback(ctx)

	var lockedListingID string

	err = tx.QueryRow(ctx, `
		SELECT id
		FROM listings
		WHERE id = $1
		AND author_id = $2
		AND deleted_at IS NULL
		AND status = 'active'
		FOR UPDATE
	`, listingID, authorID).Scan(&lockedListingID)
	if err != nil {
		return ListingPhoto{}, err
	}

	var position int16
	err = tx.QueryRow(ctx, `
		SELECT available_position::smallint
		FROM generate_series(1,3) AS available_position
		WHERE NOT EXISTS (
			SELECT 1
			FROM listing_photos
			WHERE listing_id = $1 AND position = available_position
		)
		ORDER BY available_position
		LIMIT 1;
	`, listingID).Scan(&position)
	if errors.Is(err, pgx.ErrNoRows) {
		return ListingPhoto{}, ErrListingPhotoLimitReached
	}
	if err != nil {
		return ListingPhoto{}, err
	}

	var createdListingPhoto ListingPhoto
	err = tx.QueryRow(ctx, `
		INSERT INTO listing_photos (listing_id, storage_key, position)
		VALUES ($1, $2, $3)
		RETURNING
			id,
			storage_key,
			position
	`, listingID, storageKey, position).Scan(
		&createdListingPhoto.ID,
		&createdListingPhoto.StorageKey,
		&createdListingPhoto.Position,
	)
	if err != nil {
		return ListingPhoto{}, err
	}
	return createdListingPhoto, tx.Commit(ctx)
}

func (r *Repository) FindListingPhotosByListingIDs(ctx context.Context, listingIDs []string) ([]ListingPhoto, error) {
	if len(listingIDs) == 0 {
		return []ListingPhoto{}, nil
	}
	parsedListingsIDs := make([]pgtype.UUID, 0, len(listingIDs))

	for _, listingID := range listingIDs {
		var parsedID pgtype.UUID

		if err := parsedID.Scan(listingID); err != nil {
			return nil, err
		}

		parsedListingsIDs = append(parsedListingsIDs, parsedID)
	}

	rows, err := r.db.Query(ctx, `
		SELECT id, listing_id, storage_key, position
		FROM listing_photos
		WHERE listing_id = ANY($1)
		ORDER BY listing_id, position
	`, parsedListingsIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	photos := make([]ListingPhoto, 0)

	for rows.Next() {
		var listingPhoto ListingPhoto

		if err := rows.Scan(&listingPhoto.ID, &listingPhoto.ListingID, &listingPhoto.StorageKey, &listingPhoto.Position); err != nil {
			return nil, err
		}

		photos = append(photos, listingPhoto)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return photos, nil
}

func (r *Repository) FindListingPhotoByIDAndOwner(ctx context.Context, listingID string, photoID string, authorID string) (ListingPhoto, error) {
	var photo ListingPhoto

	err := r.db.QueryRow(ctx, `
		SELECT
			photo.id,
			photo.listing_id,
			photo.storage_key,
			photo.position
		FROM listing_photos AS photo
		JOIN listings AS listing
		ON listing.id = photo.listing_id
		WHERE photo.id = $1
		AND photo.listing_id = $2
		AND listing.author_id = $3
		AND listing.status = 'active'
		AND listing.deleted_at IS NULL;
	`, photoID, listingID, authorID).Scan(
		&photo.ID,
		&photo.ListingID,
		&photo.StorageKey,
		&photo.Position,
	)

	if err != nil {
		return ListingPhoto{}, err
	}

	return photo, nil
}

func (r *Repository) DeleteListingPhoto(ctx context.Context, listingID string, photoID string) error {
	commandTag, err := r.db.Exec(ctx, `
		DELETE FROM listing_photos
		WHERE listing_id = $1 AND id = $2
	`, listingID, photoID)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
