package postgresrepo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"auth/internal/models/records"
)

type EmailChangeRequestRepo struct {
	*Repo
}

func NewEmailChangeRequestRepo(repo *Repo) *EmailChangeRequestRepo {
	return &EmailChangeRequestRepo{Repo: repo}
}

func (r *EmailChangeRequestRepo) Create(ctx context.Context, req *records.EmailChangeRequest) error {
	_, err := r.Exec(ctx, `
		insert into email_change_requests (id, identity_id, new_email, status, token_hash, expires_at, consumed_at, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, req.ID, req.IdentityID, req.NewEmail, req.Status, req.TokenHash, req.ExpiresAt, req.ConsumedAt, req.CreatedAt, req.UpdatedAt)
	return err
}

func (r *EmailChangeRequestRepo) GetByID(ctx context.Context, id uuid.UUID) (*records.EmailChangeRequest, error) {
	return r.scanEmailChangeRequest(r.QueryRow(ctx, `
		select id, identity_id, new_email, status, token_hash, expires_at, consumed_at, created_at, updated_at
		from email_change_requests
		where id = $1
	`, id))
}

func (r *EmailChangeRequestRepo) GetActiveByIdentityIDAndStatus(ctx context.Context, identityID uuid.UUID, status string) (*records.EmailChangeRequest, error) {
	return r.scanEmailChangeRequest(r.QueryRow(ctx, `
		select id, identity_id, new_email, status, token_hash, expires_at, consumed_at, created_at, updated_at
		from email_change_requests
		where identity_id = $1 and status = $2 and expires_at > now() and consumed_at is null
		order by created_at desc
		limit 1
	`, identityID, status))
}

func (r *EmailChangeRequestRepo) GetByTokenHash(ctx context.Context, hash string) (*records.EmailChangeRequest, error) {
	return r.scanEmailChangeRequest(r.QueryRow(ctx, `
		select id, identity_id, new_email, status, token_hash, expires_at, consumed_at, created_at, updated_at
		from email_change_requests
		where token_hash = $1 and consumed_at is null and expires_at > now()
	`, hash))
}

func (r *EmailChangeRequestRepo) UpdateTokenHash(ctx context.Context, id uuid.UUID, tokenHash string) error {
	_, err := r.Exec(ctx, `
		update email_change_requests
		set token_hash = $2, updated_at = now()
		where id = $1
	`, id, tokenHash)
	return err
}

func (r *EmailChangeRequestRepo) UpdateNewEmailAndStatus(ctx context.Context, id uuid.UUID, newEmail string, status string) error {
	_, err := r.Exec(ctx, `
		update email_change_requests
		set new_email = $2, status = $3, updated_at = now()
		where id = $1
	`, id, newEmail, status)
	return err
}

func (r *EmailChangeRequestRepo) SetStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.Exec(ctx, `
		update email_change_requests
		set status = $2, updated_at = now()
		where id = $1
	`, id, status)
	return err
}

func (r *EmailChangeRequestRepo) Consume(ctx context.Context, id uuid.UUID) error {
	tag, err := r.Exec(ctx, `
		update email_change_requests
		set consumed_at = now(), updated_at = now()
		where id = $1 and consumed_at is null and expires_at > now()
	`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return err
}

func (r *EmailChangeRequestRepo) scanEmailChangeRequest(row rowScanner) (*records.EmailChangeRequest, error) {
	req := new(records.EmailChangeRequest)
	err := row.Scan(
		&req.ID,
		&req.IdentityID,
		&req.NewEmail,
		&req.Status,
		&req.TokenHash,
		&req.ExpiresAt,
		&req.ConsumedAt,
		&req.CreatedAt,
		&req.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return req, nil
}
