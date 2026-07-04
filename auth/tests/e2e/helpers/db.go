package e2e_test_helpers

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IdentityRow struct {
	ID            string
	Email         string
	EmailVerified bool
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CredentialRow struct {
	IdentityID        string
	PasswordHash      string
	PasswordChangedAt time.Time
}

type SessionRow struct {
	ID               string
	IdentityID       string
	RefreshTokenHash string
	UserAgent        string
	ExpiresAt        time.Time
	RevokedAt        *time.Time
}

type VerificationCodeRow struct {
	ID            string
	IdentityID    *string
	Email         *string
	Purpose       string
	CodeHash      string
	AttemptsCount int32
	MaxAttempts   int32
	ConsumedAt    *time.Time
}

type AuthEventRow struct {
	ID         string
	IdentityID *string
	EventType  string
	CreatedAt  time.Time
}

type PasswordResetTokenRow struct {
	ID         string
	IdentityID string
	TokenHash  string
	ConsumedAt *time.Time
}

type PasswordChangeTokenRow struct {
	ID         string
	IdentityID string
	TokenHash  string
	ConsumedAt *time.Time
}

type EmailChangeRequestRow struct {
	ID         string
	IdentityID string
	NewEmail   string
	Status     string
	TokenHash  *string
	ConsumedAt *time.Time
}

func GetIdentityByEmail(t *testing.T, pool *pgxpool.Pool, email string) IdentityRow {
	t.Helper()
	var row IdentityRow
	err := pool.QueryRow(context.Background(),
		`SELECT id, email, email_verified, status, created_at, updated_at
		 FROM identities WHERE email = $1`, email,
	).Scan(&row.ID, &row.Email, &row.EmailVerified, &row.Status, &row.CreatedAt, &row.UpdatedAt)
	if err != nil {
		t.Fatalf("GetIdentityByEmail(%s): %v", email, err)
	}
	return row
}

func GetIdentityByID(t *testing.T, pool *pgxpool.Pool, id string) IdentityRow {
	t.Helper()
	var row IdentityRow
	err := pool.QueryRow(context.Background(),
		`SELECT id, email, email_verified, status, created_at, updated_at
		 FROM identities WHERE id = $1`, id,
	).Scan(&row.ID, &row.Email, &row.EmailVerified, &row.Status, &row.CreatedAt, &row.UpdatedAt)
	if err != nil {
		t.Fatalf("GetIdentityByID(%s): %v", id, err)
	}
	return row
}

func GetCredentialsByIdentityID(t *testing.T, pool *pgxpool.Pool, identityID string) CredentialRow {
	t.Helper()
	var row CredentialRow
	err := pool.QueryRow(context.Background(),
		`SELECT identity_id, password_hash, password_changed_at
		 FROM credentials WHERE identity_id = $1`, identityID,
	).Scan(&row.IdentityID, &row.PasswordHash, &row.PasswordChangedAt)
	if err != nil {
		t.Fatalf("GetCredentialsByIdentityID(%s): %v", identityID, err)
	}
	return row
}

func GetSessionsByIdentityID(t *testing.T, pool *pgxpool.Pool, identityID string) []SessionRow {
	t.Helper()
	rows, err := pool.Query(context.Background(),
		`SELECT id, identity_id, refresh_token_hash, user_agent, expires_at, revoked_at
		 FROM sessions WHERE identity_id = $1 ORDER BY created_at`, identityID,
	)
	if err != nil {
		t.Fatalf("GetSessionsByIdentityID query: %v", err)
	}
	defer rows.Close()

	var result []SessionRow
	for rows.Next() {
		var r SessionRow
		if err := rows.Scan(&r.ID, &r.IdentityID, &r.RefreshTokenHash, &r.UserAgent, &r.ExpiresAt, &r.RevokedAt); err != nil {
			t.Fatalf("GetSessionsByIdentityID scan: %v", err)
		}
		result = append(result, r)
	}
	return result
}

func GetVerificationCodesByIdentityID(t *testing.T, pool *pgxpool.Pool, identityID string) []VerificationCodeRow {
	t.Helper()
	rows, err := pool.Query(context.Background(),
		`SELECT id, identity_id, email, purpose, code_hash, attempts_count, max_attempts, consumed_at
		 FROM verification_codes WHERE identity_id = $1 ORDER BY created_at`, identityID,
	)
	if err != nil {
		t.Fatalf("GetVerificationCodesByIdentityID query: %v", err)
	}
	defer rows.Close()

	var result []VerificationCodeRow
	for rows.Next() {
		var r VerificationCodeRow
		if err := rows.Scan(&r.ID, &r.IdentityID, &r.Email, &r.Purpose, &r.CodeHash, &r.AttemptsCount, &r.MaxAttempts, &r.ConsumedAt); err != nil {
			t.Fatalf("GetVerificationCodesByIdentityID scan: %v", err)
		}
		result = append(result, r)
	}
	return result
}

func GetVerificationCodesByEmailAndPurpose(t *testing.T, pool *pgxpool.Pool, email, purpose string) []VerificationCodeRow {
	t.Helper()
	rows, err := pool.Query(context.Background(),
		`SELECT id, identity_id, email, purpose, code_hash, attempts_count, max_attempts, consumed_at
		 FROM verification_codes WHERE email = $1 AND purpose = $2 ORDER BY created_at`, email, purpose,
	)
	if err != nil {
		t.Fatalf("GetVerificationCodesByEmailAndPurpose query: %v", err)
	}
	defer rows.Close()

	var result []VerificationCodeRow
	for rows.Next() {
		var r VerificationCodeRow
		if err := rows.Scan(&r.ID, &r.IdentityID, &r.Email, &r.Purpose, &r.CodeHash, &r.AttemptsCount, &r.MaxAttempts, &r.ConsumedAt); err != nil {
			t.Fatalf("GetVerificationCodesByEmailAndPurpose scan: %v", err)
		}
		result = append(result, r)
	}
	return result
}

func GetAuthEventsByIdentityID(t *testing.T, pool *pgxpool.Pool, identityID string) []AuthEventRow {
	t.Helper()
	rows, err := pool.Query(context.Background(),
		`SELECT id, identity_id, event_type, created_at
		 FROM auth_events WHERE identity_id = $1 ORDER BY created_at`, identityID,
	)
	if err != nil {
		t.Fatalf("GetAuthEventsByIdentityID query: %v", err)
	}
	defer rows.Close()

	var result []AuthEventRow
	for rows.Next() {
		var r AuthEventRow
		if err := rows.Scan(&r.ID, &r.IdentityID, &r.EventType, &r.CreatedAt); err != nil {
			t.Fatalf("GetAuthEventsByIdentityID scan: %v", err)
		}
		result = append(result, r)
	}
	return result
}

func GetPasswordResetTokensByIdentityID(t *testing.T, pool *pgxpool.Pool, identityID string) []PasswordResetTokenRow {
	t.Helper()
	rows, err := pool.Query(context.Background(),
		`SELECT id, identity_id, token_hash, consumed_at
		 FROM password_reset_tokens WHERE identity_id = $1 ORDER BY created_at`, identityID,
	)
	if err != nil {
		t.Fatalf("GetPasswordResetTokensByIdentityID query: %v", err)
	}
	defer rows.Close()

	var result []PasswordResetTokenRow
	for rows.Next() {
		var r PasswordResetTokenRow
		if err := rows.Scan(&r.ID, &r.IdentityID, &r.TokenHash, &r.ConsumedAt); err != nil {
			t.Fatalf("GetPasswordResetTokensByIdentityID scan: %v", err)
		}
		result = append(result, r)
	}
	return result
}

func GetPasswordChangeTokensByIdentityID(t *testing.T, pool *pgxpool.Pool, identityID string) []PasswordChangeTokenRow {
	t.Helper()
	rows, err := pool.Query(context.Background(),
		`SELECT id, identity_id, token_hash, consumed_at
		 FROM password_change_tokens WHERE identity_id = $1 ORDER BY created_at`, identityID,
	)
	if err != nil {
		t.Fatalf("GetPasswordChangeTokensByIdentityID query: %v", err)
	}
	defer rows.Close()

	var result []PasswordChangeTokenRow
	for rows.Next() {
		var r PasswordChangeTokenRow
		if err := rows.Scan(&r.ID, &r.IdentityID, &r.TokenHash, &r.ConsumedAt); err != nil {
			t.Fatalf("GetPasswordChangeTokensByIdentityID scan: %v", err)
		}
		result = append(result, r)
	}
	return result
}

func GetEmailChangeRequestsByIdentityID(t *testing.T, pool *pgxpool.Pool, identityID string) []EmailChangeRequestRow {
	t.Helper()
	rows, err := pool.Query(context.Background(),
		`SELECT id, identity_id, new_email, status, token_hash, consumed_at
		 FROM email_change_requests WHERE identity_id = $1 ORDER BY created_at`, identityID,
	)
	if err != nil {
		t.Fatalf("GetEmailChangeRequestsByIdentityID query: %v", err)
	}
	defer rows.Close()

	var result []EmailChangeRequestRow
	for rows.Next() {
		var r EmailChangeRequestRow
		if err := rows.Scan(&r.ID, &r.IdentityID, &r.NewEmail, &r.Status, &r.TokenHash, &r.ConsumedAt); err != nil {
			t.Fatalf("GetEmailChangeRequestsByIdentityID scan: %v", err)
		}
		result = append(result, r)
	}
	return result
}

func GetAuthEventsByType(t *testing.T, pool *pgxpool.Pool, identityID, eventType string) []AuthEventRow {
	t.Helper()
	rows, err := pool.Query(context.Background(),
		`SELECT id, identity_id, event_type, created_at
		 FROM auth_events WHERE identity_id = $1 AND event_type = $2 ORDER BY created_at`, identityID, eventType,
	)
	if err != nil {
		t.Fatalf("GetAuthEventsByType query: %v", err)
	}
	defer rows.Close()

	var result []AuthEventRow
	for rows.Next() {
		var r AuthEventRow
		if err := rows.Scan(&r.ID, &r.IdentityID, &r.EventType, &r.CreatedAt); err != nil {
			t.Fatalf("GetAuthEventsByType scan: %v", err)
		}
		result = append(result, r)
	}
	return result
}

func CountAuthEventsByIdentityID(t *testing.T, pool *pgxpool.Pool, identityID string) int {
	t.Helper()
	var count int
	err := pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM auth_events WHERE identity_id = $1`, identityID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("CountAuthEventsByIdentityID(%s): %v", identityID, err)
	}
	return count
}

func RequireNoRows(t *testing.T, pool *pgxpool.Pool, query string, args ...any) {
	t.Helper()
	var discard int
	err := pool.QueryRow(context.Background(), query, args...).Scan(&discard)
	if err == nil {
		t.Fatalf("expected no rows for query, but got one")
	} else if err != pgx.ErrNoRows {
		t.Fatalf("expected pgx.ErrNoRows, got: %v", err)
	}
}
