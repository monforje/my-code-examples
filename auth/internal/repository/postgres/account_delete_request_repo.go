package postgresrepo

import (
	"context"

	"github.com/google/uuid"

	"auth/internal/models/records"
)

type AccountDeleteRequestRepo struct {
	*Repo
}

func NewAccountDeleteRequestRepo(repo *Repo) *AccountDeleteRequestRepo {
	return &AccountDeleteRequestRepo{Repo: repo}
}

func (r *AccountDeleteRequestRepo) Create(ctx context.Context, req *records.AccountDeleteRequest) error {
	_, err := r.Exec(ctx, `
		insert into account_delete_requests (id, identity_id, status, expires_at, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6)
	`, req.ID, req.IdentityID, req.Status, req.ExpiresAt, req.CreatedAt, req.UpdatedAt)
	return err
}

func (r *AccountDeleteRequestRepo) GetByID(ctx context.Context, id uuid.UUID) (*records.AccountDeleteRequest, error) {
	return r.scanAccountDeleteRequest(r.QueryRow(ctx, `
		select id, identity_id, status, expires_at, created_at, updated_at
		from account_delete_requests
		where id = $1
	`, id))
}

func (r *AccountDeleteRequestRepo) GetActiveByIdentityID(ctx context.Context, identityID uuid.UUID) (*records.AccountDeleteRequest, error) {
	return r.scanAccountDeleteRequest(r.QueryRow(ctx, `
		select id, identity_id, status, expires_at, created_at, updated_at
		from account_delete_requests
		where identity_id = $1 and status = 'pending' and expires_at > now()
		order by created_at desc
		limit 1
	`, identityID))
}

func (r *AccountDeleteRequestRepo) SetStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.Exec(ctx, `
		update account_delete_requests
		set status = $2, updated_at = now()
		where id = $1
	`, id, status)
	return err
}

func (r *AccountDeleteRequestRepo) scanAccountDeleteRequest(row rowScanner) (*records.AccountDeleteRequest, error) {
	req := new(records.AccountDeleteRequest)
	err := row.Scan(
		&req.ID,
		&req.IdentityID,
		&req.Status,
		&req.ExpiresAt,
		&req.CreatedAt,
		&req.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return req, nil
}
