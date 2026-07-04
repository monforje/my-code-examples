package postgresrepo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"auth/internal/models/records"
)

type PasswordChangeTokenRepo struct {
	*Repo
}

func NewPasswordChangeTokenRepo(repo *Repo) *PasswordChangeTokenRepo {
	return &PasswordChangeTokenRepo{Repo: repo}
}

func (r *PasswordChangeTokenRepo) Create(ctx context.Context, token *records.PasswordChangeToken) error {
	_, err := r.Exec(ctx, `
		insert into password_change_tokens (id, identity_id, token_hash, expires_at, consumed_at, created_at)
		values ($1, $2, $3, $4, $5, $6)
	`, token.ID, token.IdentityID, token.TokenHash, token.ExpiresAt, token.ConsumedAt, token.CreatedAt)
	return err
}

func (r *PasswordChangeTokenRepo) GetByTokenHash(ctx context.Context, hash string) (*records.PasswordChangeToken, error) {
	return r.scanPasswordChangeToken(r.QueryRow(ctx, `
		select id, identity_id, token_hash, expires_at, consumed_at, created_at
		from password_change_tokens
		where token_hash = $1 and consumed_at is null and expires_at > now()
		order by created_at desc
		limit 1
	`, hash))
}

func (r *PasswordChangeTokenRepo) Consume(ctx context.Context, id uuid.UUID) error {
	tag, err := r.Exec(ctx, `
		update password_change_tokens
		set consumed_at = now()
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

func (r *PasswordChangeTokenRepo) DeleteExpired(ctx context.Context) error {
	_, err := r.Exec(ctx, `delete from password_change_tokens where expires_at <= now()`)
	return err
}

func (r *PasswordChangeTokenRepo) scanPasswordChangeToken(row rowScanner) (*records.PasswordChangeToken, error) {
	token := new(records.PasswordChangeToken)
	err := row.Scan(
		&token.ID,
		&token.IdentityID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.ConsumedAt,
		&token.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return token, nil
}
