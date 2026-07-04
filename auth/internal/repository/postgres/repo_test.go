package postgresrepo_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"auth/internal/config"
	"auth/internal/models/records"
	postgresrepo "auth/internal/repository/postgres"
)

func newTestDB(t *testing.T) *postgresrepo.Repo {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		cfg := config.PGConfig{
			Host:     "localhost",
			Port:     "5433",
			User:     "admin",
			Password: "admin",
			DB:       "auth_db",
			SSLMode:  "disable",
		}
		dsn = cfg.DSN()
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		t.Fatalf("failed to ping test database: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
	})

	return postgresrepo.New(pool)
}

func cleanupTable(t *testing.T, repo *postgresrepo.Repo, table string) {
	t.Helper()
	_, err := repo.Exec(context.Background(), "TRUNCATE "+table+" CASCADE")
	if err != nil {
		t.Fatalf("failed to truncate %s: %v", table, err)
	}
}

func TestRepo_WithTx(t *testing.T) {
	repo := newTestDB(t)
	cleanupTable(t, repo, "identities")

	ctx := context.Background()
	identityRepo := postgresrepo.NewIdentityRepo(repo)
	committedID := uuid.New()
	now := time.Now()

	err := repo.WithTx(ctx, func(txRepo *postgresrepo.Repo) error {
		return postgresrepo.NewIdentityRepo(txRepo).Create(ctx, &records.Identity{
			ID:        committedID,
			Email:     committedID.String() + "@example.com",
			Status:    "active",
			CreatedAt: now,
			UpdatedAt: now,
		})
	})
	if err != nil {
		t.Fatalf("WithTx() commit error = %v", err)
	}

	if _, err := identityRepo.GetByID(ctx, committedID); err != nil {
		t.Fatalf("GetByID() committed identity error = %v", err)
	}

	rolledBackID := uuid.New()
	wantErr := errors.New("rollback")
	err = repo.WithTx(ctx, func(txRepo *postgresrepo.Repo) error {
		if err := postgresrepo.NewIdentityRepo(txRepo).Create(ctx, &records.Identity{
			ID:        rolledBackID,
			Email:     rolledBackID.String() + "@example.com",
			Status:    "active",
			CreatedAt: now,
			UpdatedAt: now,
		}); err != nil {
			return err
		}

		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("WithTx() rollback error = %v, want %v", err, wantErr)
	}

	if _, err := identityRepo.GetByID(ctx, rolledBackID); err == nil {
		t.Fatal("GetByID() rolled back identity error = nil")
	}
}

func createTestIdentity(t *testing.T, repo *postgresrepo.Repo) uuid.UUID {
	t.Helper()

	identityRepo := postgresrepo.NewIdentityRepo(repo)
	id := uuid.New()
	now := time.Now()

	err := identityRepo.Create(context.Background(), &records.Identity{
		ID:        id,
		Email:     id.String() + "@example.com",
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("failed to create test identity: %v", err)
	}

	return id
}
