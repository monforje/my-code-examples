package postgresrepo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"users/internal/models/records"
)

var ErrUserProfileNotFound = errors.New("user profile not found")

type UserProfileRepo struct {
	*Repo
}

func NewUserProfileRepo(repo *Repo) *UserProfileRepo {
	return &UserProfileRepo{
		Repo: repo,
	}
}

func (r *UserProfileRepo) Create(ctx context.Context, profile *records.UserProfile) error {
	_, err := r.Exec(ctx, `
		INSERT INTO user_profiles (id, identity_id, email, display_name, bio, avatar_url, avatar_object_key, status, email_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, profile.ID, profile.IdentityID, profile.Email, profile.DisplayName, profile.BIO, profile.AvatarURL, profile.AvatarObjectKey, profile.Status, profile.EmailVerified, profile.CreatedAt, profile.UpdatedAt)
	return err
}

func (r *UserProfileRepo) GetByID(ctx context.Context, id uuid.UUID) (*records.UserProfile, error) {
	var p records.UserProfile
	err := r.QueryRow(ctx, `
		SELECT id, identity_id, email, display_name, bio, avatar_url, avatar_object_key, status, email_verified, created_at, updated_at, deleted_at
		FROM user_profiles
		WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(&p.ID, &p.IdentityID, &p.Email, &p.DisplayName, &p.BIO, &p.AvatarURL, &p.AvatarObjectKey, &p.Status, &p.EmailVerified, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserProfileNotFound
	}
	return &p, err
}

func (r *UserProfileRepo) GetByEmail(ctx context.Context, email string) (*records.UserProfile, error) {
	var p records.UserProfile
	err := r.QueryRow(ctx, `
		SELECT id, identity_id, email, display_name, bio, avatar_url, avatar_object_key, status, email_verified, created_at, updated_at, deleted_at
		FROM user_profiles
		WHERE email = $1 AND deleted_at IS NULL
	`, email).Scan(&p.ID, &p.IdentityID, &p.Email, &p.DisplayName, &p.BIO, &p.AvatarURL, &p.AvatarObjectKey, &p.Status, &p.EmailVerified, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserProfileNotFound
	}
	return &p, err
}

func (r *UserProfileRepo) GetByIdentityID(ctx context.Context, identityID uuid.UUID) (*records.UserProfile, error) {
	var p records.UserProfile
	err := r.QueryRow(ctx, `
		SELECT id, identity_id, email, display_name, bio, avatar_url, avatar_object_key, status, email_verified, created_at, updated_at, deleted_at
		FROM user_profiles
		WHERE identity_id = $1 AND deleted_at IS NULL
	`, identityID).Scan(&p.ID, &p.IdentityID, &p.Email, &p.DisplayName, &p.BIO, &p.AvatarURL, &p.AvatarObjectKey, &p.Status, &p.EmailVerified, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserProfileNotFound
	}
	return &p, err
}

func (r *UserProfileRepo) ExistsByDisplayName(ctx context.Context, displayName string) (bool, error) {
	var exists bool
	err := r.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM user_profiles
			WHERE display_name = $1 AND deleted_at IS NULL
		)
	`, displayName).Scan(&exists)
	return exists, err
}

func (r *UserProfileRepo) Update(ctx context.Context, profile *records.UserProfile) error {
	tag, err := r.Exec(ctx, `
		UPDATE user_profiles
		SET display_name = $2, bio = $3, avatar_url = $4, avatar_object_key = $5, status = $6, email_verified = $7, updated_at = $8
		WHERE id = $1 AND deleted_at IS NULL
	`, profile.ID, profile.DisplayName, profile.BIO, profile.AvatarURL, profile.AvatarObjectKey, profile.Status, profile.EmailVerified, profile.UpdatedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrUserProfileNotFound
	}
	return nil
}

func (r *UserProfileRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.Exec(ctx, `
		UPDATE user_profiles
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrUserProfileNotFound
	}
	return nil
}
