package postgresrepo_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"users/internal/config"
	postgresrepo "users/internal/repository/postgres"
)

func newTestDB(t *testing.T) *postgresrepo.Repo {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		cfg := config.PGConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "admin",
			Password: "admin",
			DB:       "users_db",
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
	runTestMigrations(t, dsn)

	t.Cleanup(func() {
		pool.Close()
	})

	return postgresrepo.New(pool)
}

func runTestMigrations(t *testing.T, dsn string) {
	t.Helper()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open migration db: %v", err)
	}
	defer db.Close()

	migrationsDir, err := testMigrationsDir()
	if err != nil {
		t.Fatalf("resolve migrations dir: %v", err)
	}
	if err := goose.UpContext(context.Background(), db, migrationsDir); err != nil {
		t.Fatalf("run migrations: %v", err)
	}
}

func testMigrationsDir() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", os.ErrNotExist
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "migrations")), nil
}

func cleanupTable(t *testing.T, repo *postgresrepo.Repo, table string) {
	t.Helper()
	_, err := repo.Exec(context.Background(), "TRUNCATE "+table+" CASCADE")
	if err != nil {
		t.Fatalf("failed to truncate %s: %v", table, err)
	}
}
