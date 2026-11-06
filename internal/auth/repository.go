package auth

import (
	"context"
	"time"

	"github.com/Dimes2212/5th_semester_course_work_backend/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindUserByEmail(ctx context.Context, email string) (user.User, error) {
	var foundUser user.User

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
		WHERE email = $1
		  AND deleted_at IS NULL
	`, email).Scan(
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

func (r *Repository) CreateUser(ctx context.Context, fullName string, email string, PasswordHash string) (user.User, error) {
	var newUser user.User

	err := r.db.QueryRow(ctx, `
		INSERT INTO users (full_name, email, password_hash)
		VALUES ($1, $2, $3)
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
	`, fullName, email, PasswordHash).Scan(
		&newUser.ID,
		&newUser.FullName,
		&newUser.Email,
		&newUser.PasswordHash,
		&newUser.Role,
		&newUser.AvatarKey,
		&newUser.CreatedAt,
		&newUser.UpdatedAt,
		&newUser.DeletedAt,
	)

	return newUser, err
}

func (r *Repository) CreateRefreshSession(ctx context.Context, session RefreshSession) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO refresh_sessions (user_id, token_hash, device_name, expires_at)
		VALUES ($1, $2, $3, $4)
	`, session.UserID, session.TokenHash, session.DeviceName, session.ExpiresAt)

	return err
}

func (r *Repository) RotateRefreshSession(
	ctx context.Context,
	oldTokenHash string,
	newTokenHash string,
	expiresAt time.Time,
) (user.User, error) {
	var foundUser user.User

	err := r.db.QueryRow(ctx, `
		WITH revoked_session AS (
			UPDATE refresh_sessions AS session
			SET revoked_at = now()
			FROM users
			WHERE session.token_hash = $1
			  AND session.revoked_at IS NULL
			  AND session.expires_at > now()
			  AND users.id = session.user_id
			  AND users.deleted_at IS NULL
			RETURNING session.user_id, session.device_name
		),
		created_session AS (
			INSERT INTO refresh_sessions (user_id, token_hash, device_name, expires_at)
			SELECT user_id, $2, device_name, $3
			FROM revoked_session
			RETURNING user_id
		)
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
		WHERE id = (
			SELECT user_id
			FROM created_session
		)
	`, oldTokenHash, newTokenHash, expiresAt).Scan(
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

func (r *Repository) RevokeRefreshSession(ctx context.Context, tokenHash string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE refresh_sessions
		SET revoked_at = now()
		WHERE token_hash = $1
		  AND revoked_at IS NULL
	`, tokenHash)

	return err
}
