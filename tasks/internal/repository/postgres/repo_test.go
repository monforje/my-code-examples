package postgresrepo_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"tasks/internal/config"
	postgresrepo "tasks/internal/repository/postgres"
)

func newTestDB(t *testing.T) *postgresrepo.Repo {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		cfg := config.PGConfig{
			Host:     "localhost",
			Port:     "5435",
			User:     "admin",
			Password: "admin",
			DB:       "tasks_db",
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
