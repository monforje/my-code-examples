package postgresrepo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"auth/internal/models/records"
)

type VerificationCodeRepo struct {
	*Repo
}

func NewVerificationCodeRepo(repo *Repo) *VerificationCodeRepo {
	return &VerificationCodeRepo{Repo: repo}
}

func (r *VerificationCodeRepo) Create(ctx context.Context, code *records.VerificationCode) error {
	_, err := r.Exec(ctx, `
		insert into verification_codes (id, identity_id, email, purpose, code_hash, attempts_count, max_attempts, expires_at, consumed_at, created_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, code.ID, code.IdentityID, code.Email, code.Purpose, code.CodeHash, code.AttemptsCount, code.MaxAttempts, code.ExpiresAt, code.ConsumedAt, code.CreatedAt)
	return err
}

func (r *VerificationCodeRepo) GetByID(ctx context.Context, id uuid.UUID) (*records.VerificationCode, error) {
	return r.scanVerificationCode(r.QueryRow(ctx, `
		select id, identity_id, email, purpose, code_hash, attempts_count, max_attempts, expires_at, consumed_at, created_at
		from verification_codes
		where id = $1
	`, id))
}

func (r *VerificationCodeRepo) GetActiveByEmailAndPurpose(ctx context.Context, email, purpose string) (*records.VerificationCode, error) {
	return r.scanVerificationCode(r.QueryRow(ctx, `
		select id, identity_id, email, purpose, code_hash, attempts_count, max_attempts, expires_at, consumed_at, created_at
		from verification_codes
		where email = $1 and purpose = $2 and consumed_at is null and expires_at > now()
		order by created_at desc
		limit 1
	`, email, purpose))
}

func (r *VerificationCodeRepo) GetActiveByIdentityIDAndPurpose(ctx context.Context, identityID uuid.UUID, purpose string) (*records.VerificationCode, error) {
	return r.scanVerificationCode(r.QueryRow(ctx, `
		select id, identity_id, email, purpose, code_hash, attempts_count, max_attempts, expires_at, consumed_at, created_at
		from verification_codes
		where identity_id = $1 and purpose = $2 and consumed_at is null and expires_at > now()
		order by created_at desc
		limit 1
	`, identityID, purpose))
}

func (r *VerificationCodeRepo) IncrementAttempts(ctx context.Context, id uuid.UUID) error {
	_, err := r.Exec(ctx, `
		update verification_codes
		set attempts_count = attempts_count + 1
		where id = $1
	`, id)
	return err
}

func (r *VerificationCodeRepo) Consume(ctx context.Context, id uuid.UUID) error {
	tag, err := r.Exec(ctx, `
		update verification_codes
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

func (r *VerificationCodeRepo) DeleteExpired(ctx context.Context) error {
	_, err := r.Exec(ctx, `delete from verification_codes where expires_at <= now()`)
	return err
}

func (r *VerificationCodeRepo) scanVerificationCode(row rowScanner) (*records.VerificationCode, error) {
	code := new(records.VerificationCode)
	err := row.Scan(
		&code.ID,
		&code.IdentityID,
		&code.Email,
		&code.Purpose,
		&code.CodeHash,
		&code.AttemptsCount,
		&code.MaxAttempts,
		&code.ExpiresAt,
		&code.ConsumedAt,
		&code.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return code, nil
}
