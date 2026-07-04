package postgresrepo

import (
	"context"

	"github.com/google/uuid"

	"auth/internal/models/records"
)

type IdentityRepo struct {
	*Repo
}

func NewIdentityRepo(repo *Repo) *IdentityRepo {
	return &IdentityRepo{Repo: repo}
}

func (r *IdentityRepo) Create(ctx context.Context, identity *records.Identity) error {
	_, err := r.Exec(ctx, `
		insert into identities (id, email, email_verified, status, created_at, updated_at, deleted_at)
		values ($1, $2, $3, $4, $5, $6, $7)
	`, identity.ID, identity.Email, identity.EmailVerified, identity.Status, identity.CreatedAt, identity.UpdatedAt, identity.DeletedAt)
	return err
}

func (r *IdentityRepo) GetByID(ctx context.Context, id uuid.UUID) (*records.Identity, error) {
	identity := new(records.Identity)
	err := r.QueryRow(ctx, `
		select id, email, email_verified, status, created_at, updated_at, deleted_at
		from identities
		where id = $1
	`, id).Scan(
		&identity.ID,
		&identity.Email,
		&identity.EmailVerified,
		&identity.Status,
		&identity.CreatedAt,
		&identity.UpdatedAt,
		&identity.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return identity, nil
}

func (r *IdentityRepo) GetByEmail(ctx context.Context, email string) (*records.Identity, error) {
	identity := new(records.Identity)
	err := r.QueryRow(ctx, `
		select id, email, email_verified, status, created_at, updated_at, deleted_at
		from identities
		where email = $1
	`, email).Scan(
		&identity.ID,
		&identity.Email,
		&identity.EmailVerified,
		&identity.Status,
		&identity.CreatedAt,
		&identity.UpdatedAt,
		&identity.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return identity, nil
}

func (r *IdentityRepo) Update(ctx context.Context, identity *records.Identity) error {
	_, err := r.Exec(ctx, `
		update identities
		set email = $2, email_verified = $3, status = $4, updated_at = $5, deleted_at = $6
		where id = $1
	`, identity.ID, identity.Email, identity.EmailVerified, identity.Status, identity.UpdatedAt, identity.DeletedAt)
	return err
}

func (r *IdentityRepo) SetEmailVerified(ctx context.Context, id uuid.UUID) error {
	_, err := r.Exec(ctx, `
		update identities
		set email_verified = true, updated_at = now()
		where id = $1
	`, id)
	return err
}

func (r *IdentityRepo) SetStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.Exec(ctx, `
		update identities
		set status = $2, updated_at = now()
		where id = $1
	`, id, status)
	return err
}

func (r *IdentityRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.Exec(ctx, `
		update identities
		set status = 'deleted', deleted_at = now(), updated_at = now()
		where id = $1
	`, id)
	return err
}
