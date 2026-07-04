package postgresrepo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"auth/internal/models/records"
)

type PasswordResetTokenRepo struct {
	*Repo
}

func NewPasswordResetTokenRepo(repo *Repo) *PasswordResetTokenRepo {
	return &PasswordResetTokenRepo{Repo: repo}
}

func (r *PasswordResetTokenRepo) Create(ctx context.Context, token *records.PasswordResetToken) error {
	_, err := r.Exec(ctx, `
		insert into password_reset_tokens (id, identity_id, token_hash, expires_at, consumed_at, created_at)
		values ($1, $2, $3, $4, $5, $6)
	`, token.ID, token.IdentityID, token.TokenHash, token.ExpiresAt, token.ConsumedAt, token.CreatedAt)
	return err
}

func (r *PasswordResetTokenRepo) GetByID(ctx context.Context, id uuid.UUID) (*records.PasswordResetToken, error) {
	return r.scanPasswordResetToken(r.QueryRow(ctx, `
		select id, identity_id, token_hash, expires_at, consumed_at, created_at
		from password_reset_tokens
		where id = $1
	`, id))
}

func (r *PasswordResetTokenRepo) GetByTokenHash(ctx context.Context, hash string) (*records.PasswordResetToken, error) {
	return r.scanPasswordResetToken(r.QueryRow(ctx, `
		select id, identity_id, token_hash, expires_at, consumed_at, created_at
		from password_reset_tokens
		where token_hash = $1 and consumed_at is null and expires_at > now()
		order by created_at desc
		limit 1
	`, hash))
}

func (r *PasswordResetTokenRepo) Consume(ctx context.Context, id uuid.UUID) error {
	tag, err := r.Exec(ctx, `
		update password_reset_tokens
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

func (r *PasswordResetTokenRepo) DeleteExpired(ctx context.Context) error {
	_, err := r.Exec(ctx, `delete from password_reset_tokens where expires_at <= now()`)
	return err
}

func (r *PasswordResetTokenRepo) scanPasswordResetToken(row rowScanner) (*records.PasswordResetToken, error) {
	token := new(records.PasswordResetToken)
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
