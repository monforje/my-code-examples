package e2e_test_helpers

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"notifications/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	postgresUser     = "e2e"
	postgresPassword = "e2e"
	postgresDB       = "notifications_e2e"
	natsUser         = "e2e"
	natsPassword     = "e2e"
	redisPassword    = "e2e"
)

type Environment struct {
	PostgresConfig config.PGConfig
	NATSConfig     config.NATSConfig
	RedisConfig    config.RedisConfig

	postgres testcontainers.Container
	nats     testcontainers.Container
	redis    testcontainers.Container

	pgPool *pgxpool.Pool
	rdb    *redis.Client
}

func (e *Environment) PgPool() *pgxpool.Pool { return e.pgPool }
func (e *Environment) RDB() *redis.Client    { return e.rdb }

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
	if err := env.startRedis(ctx); err != nil {
		env.Shutdown(context.Background())
		return nil, err
	}
	if err := env.startNATS(ctx); err != nil {
		env.Shutdown(context.Background())
		return nil, err
	}

	return env, nil
}

func (e *Environment) Shutdown(ctx context.Context) {
	if e.rdb != nil {
		_ = e.rdb.Close()
	}
	if e.pgPool != nil {
		e.pgPool.Close()
	}
	if e.redis != nil {
		_ = e.redis.Terminate(ctx)
	}
	if e.nats != nil {
		_ = e.nats.Terminate(ctx)
	}
	if e.postgres != nil {
		_ = e.postgres.Terminate(ctx)
	}
}

func (e *Environment) Reset(ctx context.Context) error {
	_, err := e.pgPool.Exec(ctx, `TRUNCATE processed_events CASCADE`)
	if err != nil {
		return fmt.Errorf("truncate postgres: %w", err)
	}
	if err := e.rdb.FlushDB(ctx).Err(); err != nil {
		return fmt.Errorf("flush redis: %w", err)
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

func (e *Environment) startNATS(ctx context.Context) error {
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "nats:2.14-alpine3.22",
			ExposedPorts: []string{"4222/tcp"},
			Cmd:          []string{"nats-server", "--user", natsUser, "--pass", natsPassword},
			WaitingFor:   wait.ForListeningPort("4222/tcp").WithStartupTimeout(2 * time.Minute),
		},
		Started: true,
	})
	if err != nil {
		return fmt.Errorf("start nats: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return fmt.Errorf("nats host: %w", err)
	}
	port, err := container.MappedPort(ctx, "4222/tcp")
	if err != nil {
		return fmt.Errorf("nats port: %w", err)
	}

	e.nats = container
	e.NATSConfig = config.NATSConfig{
		Host:     host,
		Port:     port.Port(),
		User:     natsUser,
		Password: natsPassword,
	}

	return nil
}

func (e *Environment) startRedis(ctx context.Context) error {
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "redis:8-alpine3.21",
			ExposedPorts: []string{"6379/tcp"},
			Cmd:          []string{"redis-server", "--requirepass", redisPassword},
			WaitingFor:   wait.ForListeningPort("6379/tcp").WithStartupTimeout(2 * time.Minute),
		},
		Started: true,
	})
	if err != nil {
		return fmt.Errorf("start redis: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return fmt.Errorf("redis host: %w", err)
	}
	port, err := container.MappedPort(ctx, "6379/tcp")
	if err != nil {
		return fmt.Errorf("redis port: %w", err)
	}

	e.redis = container
	e.RedisConfig = config.RedisConfig{
		Host:     host,
		Port:     port.Port(),
		Password: redisPassword,
		DB:       0,
	}
	e.rdb = redis.NewClient(&redis.Options{Addr: e.RedisConfig.Addr(), Password: redisPassword})
	if err := e.rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping redis: %w", err)
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
