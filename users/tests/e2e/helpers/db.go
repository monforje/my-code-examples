// Package e2e_test_helpers
package e2e_test_helpers

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserProfileRow struct {
	ID              string
	IdentityID      string
	Email           string
	DisplayName     string
	Bio             string
	AvatarURL       string
	AvatarObjectKey string
	Status          string
	EmailVerified   bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

type ProcessedEventRow struct {
	EventID     string
	EventType   string
	AggregateID string
	ProcessedAt time.Time
}

type GitUserRow struct {
	ID        string
	ProfileID string
	GitToken  string
	GitURL    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func GetUserProfileByIdentityID(t *testing.T, pool *pgxpool.Pool, identityID string) UserProfileRow {
	t.Helper()
	var row UserProfileRow
	err := pool.QueryRow(context.Background(),
		`SELECT id, identity_id, email, display_name, bio, avatar_url, avatar_object_key, status, email_verified, created_at, updated_at, deleted_at
		 FROM user_profiles WHERE identity_id = $1 AND deleted_at IS NULL`, identityID,
	).Scan(&row.ID, &row.IdentityID, &row.Email, &row.DisplayName, &row.Bio, &row.AvatarURL, &row.AvatarObjectKey, &row.Status, &row.EmailVerified, &row.CreatedAt, &row.UpdatedAt, &row.DeletedAt)
	if err != nil {
		t.Fatalf("GetUserProfileByIdentityID(%s): %v", identityID, err)
	}
	return row
}

func GetUserProfileByID(t *testing.T, pool *pgxpool.Pool, id string) UserProfileRow {
	t.Helper()
	var row UserProfileRow
	err := pool.QueryRow(context.Background(),
		`SELECT id, identity_id, email, display_name, bio, avatar_url, avatar_object_key, status, email_verified, created_at, updated_at, deleted_at
		 FROM user_profiles WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&row.ID, &row.IdentityID, &row.Email, &row.DisplayName, &row.Bio, &row.AvatarURL, &row.AvatarObjectKey, &row.Status, &row.EmailVerified, &row.CreatedAt, &row.UpdatedAt, &row.DeletedAt)
	if err != nil {
		t.Fatalf("GetUserProfileByID(%s): %v", id, err)
	}
	return row
}

func GetAllUserProfiles(t *testing.T, pool *pgxpool.Pool) []UserProfileRow {
	t.Helper()
	rows, err := pool.Query(context.Background(),
		`SELECT id, identity_id, email, display_name, bio, avatar_url, avatar_object_key, status, email_verified, created_at, updated_at, deleted_at
		 FROM user_profiles WHERE deleted_at IS NULL ORDER BY created_at`,
	)
	if err != nil {
		t.Fatalf("GetAllUserProfiles query: %v", err)
	}
	defer rows.Close()

	var result []UserProfileRow
	for rows.Next() {
		var r UserProfileRow
		if err := rows.Scan(&r.ID, &r.IdentityID, &r.Email, &r.DisplayName, &r.Bio, &r.AvatarURL, &r.AvatarObjectKey, &r.Status, &r.EmailVerified, &r.CreatedAt, &r.UpdatedAt, &r.DeletedAt); err != nil {
			t.Fatalf("GetAllUserProfiles scan: %v", err)
		}
		result = append(result, r)
	}
	return result
}

func GetProcessedEventByEventID(t *testing.T, pool *pgxpool.Pool, eventID string) ProcessedEventRow {
	t.Helper()
	var row ProcessedEventRow
	err := pool.QueryRow(context.Background(),
		`SELECT event_id, event_type, aggregate_id, processed_at
		 FROM processed_events WHERE event_id = $1`, eventID,
	).Scan(&row.EventID, &row.EventType, &row.AggregateID, &row.ProcessedAt)
	if err != nil {
		t.Fatalf("GetProcessedEventByEventID(%s): %v", eventID, err)
	}
	return row
}

func GetGitUserByIdentityID(t *testing.T, pool *pgxpool.Pool, identityID string) GitUserRow {
	t.Helper()
	var row GitUserRow
	err := pool.QueryRow(context.Background(),
		`SELECT id, profile_id, git_token, git_url, created_at, updated_at
		 FROM git_users WHERE profile_id = $1`, identityID,
	).Scan(&row.ID, &row.ProfileID, &row.GitToken, &row.GitURL, &row.CreatedAt, &row.UpdatedAt)
	if err != nil {
		t.Fatalf("GetGitUserByIdentityID(%s): %v", identityID, err)
	}
	return row
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
