package e2e_test_helpers

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"users/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	postgresUser     = "e2e"
	postgresPassword = "e2e"
	postgresDB       = "users_e2e"
)

type Environment struct {
	PostgresConfig config.PGConfig
	AvatarDir      string

	postgres testcontainers.Container

	pgPool *pgxpool.Pool
}

func (e *Environment) PgPool() *pgxpool.Pool { return e.pgPool }

func StartEnvironment(ctx context.Context) (*Environment, error) {
	env := &Environment{}

	if err := env.startPostgres(ctx); err != nil {
		env.Shutdown(context.Background())
		return nil, err
	}
	if err := env.runMigrations(ctx); err != nil {
		env.Shutdown(context.Background())
		return nil, err
	}

	tmpDir, err := os.MkdirTemp("", "users-e2e-avatars-*")
	if err != nil {
		env.Shutdown(context.Background())
		return nil, fmt.Errorf("create avatar temp dir: %w", err)
	}
	env.AvatarDir = tmpDir

	return env, nil
}

func (e *Environment) Shutdown(ctx context.Context) {
	if e.pgPool != nil {
		e.pgPool.Close()
	}
	if e.postgres != nil {
		_ = e.postgres.Terminate(ctx)
	}
	if e.AvatarDir != "" {
		_ = os.RemoveAll(e.AvatarDir)
	}
}

func (e *Environment) Reset(ctx context.Context) error {
	_, err := e.pgPool.Exec(ctx, `
		TRUNCATE
			git_users,
			user_profiles,
			processed_events
		CASCADE
	`)
	if err != nil {
		return fmt.Errorf("truncate postgres: %w", err)
	}

	entries, err := os.ReadDir(e.AvatarDir)
	if err == nil {
		for _, entry := range entries {
			_ = os.RemoveAll(filepath.Join(e.AvatarDir, entry.Name()))
		}
	}

	return nil
}

func (e *Environment) startPostgres(ctx context.Context) error {
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:17-alpine3.23",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     postgresUser,
				"POSTGRES_PASSWORD": postgresPassword,
				"POSTGRES_DB":       postgresDB,
			},
			WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(2 * time.Minute),
		},
		Started: true,
	})
	if err != nil {
		return fmt.Errorf("start postgres: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return fmt.Errorf("postgres host: %w", err)
	}
	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		return fmt.Errorf("postgres port: %w", err)
	}

	e.postgres = container
	e.PostgresConfig = config.PGConfig{
		Host:     host,
		Port:     port.Port(),
		User:     postgresUser,
		Password: postgresPassword,
		DB:       postgresDB,
		SSLMode:  "disable",
	}

	e.pgPool, err = pgxpool.New(ctx, e.PostgresConfig.DSN())
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	if err := e.pgPool.Ping(ctx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}

	return nil
}

func (e *Environment) runMigrations(ctx context.Context) error {
	db, err := sql.Open("pgx", e.PostgresConfig.DSN())
	if err != nil {
		return fmt.Errorf("open migration db: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping migration db: %w", err)
	}

	migrationsDir, err := migrationsDir()
	if err != nil {
		return err
	}
	if err := goose.UpContext(ctx, db, migrationsDir); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}

func migrationsDir() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("resolve current file")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "migrations")), nil
}
