package postgresrepo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"auth/internal/models/records"
)

type SessionRepo struct {
	*Repo
}

func NewSessionRepo(repo *Repo) *SessionRepo {
	return &SessionRepo{Repo: repo}
}

func (r *SessionRepo) Create(ctx context.Context, session *records.Session) error {
	_, err := r.Exec(ctx, `
		insert into sessions (id, identity_id, refresh_token_hash, user_agent, ip_address, expires_at, revoked_at, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, session.ID, session.IdentityID, session.RefreshTokenHash, session.UserAgent, session.IPAddress, session.ExpiresAt, session.RevokedAt, session.CreatedAt, session.UpdatedAt)
	return err
}

func (r *SessionRepo) GetByID(ctx context.Context, id uuid.UUID) (*records.Session, error) {
	return r.scanSession(r.QueryRow(ctx, `
		select id, identity_id, refresh_token_hash, user_agent, ip_address, expires_at, revoked_at, created_at, updated_at
		from sessions
		where id = $1
	`, id))
}

func (r *SessionRepo) GetByRefreshTokenHash(ctx context.Context, hash string) (*records.Session, error) {
	return r.scanSession(r.QueryRow(ctx, `
		select id, identity_id, refresh_token_hash, user_agent, ip_address, expires_at, revoked_at, created_at, updated_at
		from sessions
		where refresh_token_hash = $1
	`, hash))
}

func (r *SessionRepo) GetActiveByIdentityID(ctx context.Context, identityID uuid.UUID) ([]*records.Session, error) {
	rows, err := r.Query(ctx, `
		select id, identity_id, refresh_token_hash, user_agent, ip_address, expires_at, revoked_at, created_at, updated_at
		from sessions
		where identity_id = $1 and revoked_at is null and expires_at > now()
		order by created_at desc
	`, identityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]*records.Session, 0)
	for rows.Next() {
		session, err := r.scanSession(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *SessionRepo) Revoke(ctx context.Context, id uuid.UUID) error {
	tag, err := r.Exec(ctx, `
		update sessions
		set revoked_at = now(), updated_at = now()
		where id = $1 and revoked_at is null and expires_at > now()
	`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return err
}

func (r *SessionRepo) RevokeAllByIdentityID(ctx context.Context, identityID uuid.UUID) error {
	_, err := r.Exec(ctx, `
		update sessions
		set revoked_at = now(), updated_at = now()
		where identity_id = $1 and revoked_at is null
	`, identityID)
	return err
}

func (r *SessionRepo) DeleteExpired(ctx context.Context) error {
	_, err := r.Exec(ctx, `delete from sessions where expires_at <= now()`)
	return err
}

type rowScanner interface {
	Scan(dest ...any) error
}

func (r *SessionRepo) scanSession(row rowScanner) (*records.Session, error) {
	session := new(records.Session)
	err := row.Scan(
		&session.ID,
		&session.IdentityID,
		&session.RefreshTokenHash,
		&session.UserAgent,
		&session.IPAddress,
		&session.ExpiresAt,
		&session.RevokedAt,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return session, nil
}

var _ rowScanner = pgx.Row(nil)
