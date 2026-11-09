package user

import (
	"context"

	"github.com/jackc/pgx/v5"
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

func (r *Repository) FindByID(ctx context.Context, userID string) (User, error) {
	var foundUser User

	err := r.db.QueryRow(ctx, `
		SELECT
			id,
			full_name,
			email,
			password_hash,
			role,
			avatar_key,
			created_at,
			updated_at,
			deleted_at
		FROM users
		WHERE id = $1
		  AND deleted_at IS NULL
	`, userID).Scan(
		&foundUser.ID,
		&foundUser.FullName,
		&foundUser.Email,
		&foundUser.PasswordHash,
		&foundUser.Role,
		&foundUser.AvatarKey,
		&foundUser.CreatedAt,
		&foundUser.UpdatedAt,
		&foundUser.DeletedAt,
	)

	return foundUser, err
}

func (r *Repository) UpdateFullName(ctx context.Context, userID string, fullName string) (User, error) {
	var updatedUser User

	err := r.db.QueryRow(ctx, `
		UPDATE users
		SET full_name = $1,
			updated_at = now()
		WHERE id = $2 AND deleted_at IS NULL
		RETURNING
			id,
			full_name,
			email,
			password_hash,
			role,
			avatar_key,
			created_at,
			updated_at,
			deleted_at;
	`, fullName, userID).Scan(
		&updatedUser.ID,
		&updatedUser.FullName,
		&updatedUser.Email,
		&updatedUser.PasswordHash,
		&updatedUser.Role,
		&updatedUser.AvatarKey,
		&updatedUser.CreatedAt,
		&updatedUser.UpdatedAt,
		&updatedUser.DeletedAt,
	)

	return updatedUser, err
}

func (r *Repository) DeleteByID(ctx context.Context, userID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	result, err := tx.Exec(ctx, `
		UPDATE users
		SET deleted_at = now(),
		    updated_at = now()
		WHERE id = $1
		  AND deleted_at IS NULL
	`, userID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	_, err = tx.Exec(ctx, `
		UPDATE refresh_sessions
		SET revoked_at = now()
		WHERE user_id = $1
		  AND revoked_at IS NULL
	`, userID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
