// Package rediscache
package rediscache_test

import (
	"context"
	"os"
	"testing"
	"time"

	"notifications/internal/models/domain"
	rediscache "notifications/internal/repository/redis"

	"github.com/redis/go-redis/v9"
)

func newTestRedis(t *testing.T) *redis.Client {
	t.Helper()

	addr := os.Getenv("TEST_REDIS_ADDR")
	useDefault := addr == ""
	if addr == "" {
		addr = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{Addr: addr})

	if err := client.Ping(context.Background()).Err(); err != nil {
		client.Close()
		if useDefault {
			t.Skipf("skip redis cache tests: %v", err)
		}
		t.Fatalf("failed to connect to test redis: %v", err)
	}

	t.Cleanup(func() {
		client.Close()
	})

	return client
}

func flushRedis(t *testing.T, client *redis.Client) {
	t.Helper()
	if err := client.FlushDB(context.Background()).Err(); err != nil {
		t.Fatalf("flush db: %v", err)
	}
}

func sampleReport(uid, commit string, status domain.CIReportStatus) *domain.CIReport {
	return &domain.CIReport{
		UID:       uid,
		Commit:    commit,
		Status:    status,
		ExitCode:  0,
		Stage:     "run",
		Stdout:    "build ok\ntests passed",
		Stderr:    "",
		CreatedAt: time.Now().UTC().Truncate(time.Second),
		Steps: []domain.CIReportStep{
			{Name: "build", Status: domain.StepSuccess, DurationMs: 1200},
			{Name: "test", Status: domain.StepSuccess, DurationMs: 800},
		},
	}
}

func TestReportCache_SaveAndGetLatest(t *testing.T) {
	client := newTestRedis(t)
	flushRedis(t, client)

	cache := rediscache.NewReportCache(client, time.Hour)
	ctx := context.Background()
	report := sampleReport("alice/golden-pizza-api", "abc123", domain.StatusPassed)

	// 1. Save
	if err := cache.Save(ctx, report); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// 2. GetLatest
	got, err := cache.GetLatest(ctx, "alice/golden-pizza-api")
	if err != nil {
		t.Fatalf("GetLatest() error = %v", err)
	}
	if got.UID != report.UID || got.Commit != report.Commit || got.Status != report.Status {
		t.Fatalf("got = %+v, want %+v", got, report)
	}
	if len(got.Steps) != 2 {
		t.Fatalf("steps len = %d, want 2", len(got.Steps))
	}
}

func TestReportCache_SaveAndGetByCommit(t *testing.T) {
	client := newTestRedis(t)
	flushRedis(t, client)

	cache := rediscache.NewReportCache(client, time.Hour)
	ctx := context.Background()
	report := sampleReport("alice/golden-pizza-api", "def456", domain.StatusFailed)
	report.ExitCode = 1

	if err := cache.Save(ctx, report); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := cache.GetByCommit(ctx, "alice/golden-pizza-api", "def456")
	if err != nil {
		t.Fatalf("GetByCommit() error = %v", err)
	}
	if got.Commit != "def456" || got.ExitCode != 1 {
		t.Fatalf("got = %+v", got)
	}
}

func TestReportCache_GetLatest_NotFound(t *testing.T) {
	client := newTestRedis(t)
	flushRedis(t, client)

	cache := rediscache.NewReportCache(client, time.Hour)
	ctx := context.Background()

	_, err := cache.GetLatest(ctx, "nobody/golden-pizza-api")
	if err != redis.Nil {
		t.Fatalf("error = %v, want redis.Nil", err)
	}
}

func TestReportCache_LatestOverwrite(t *testing.T) {
	client := newTestRedis(t)
	flushRedis(t, client)

	cache := rediscache.NewReportCache(client, time.Hour)
	ctx := context.Background()

	// Save first commit
	r1 := sampleReport("alice/golden-pizza-api", "aaa111", domain.StatusPassed)
	if err := cache.Save(ctx, r1); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Save second commit — should overwrite latest
	r2 := sampleReport("alice/golden-pizza-api", "bbb222", domain.StatusFailed)
	r2.ExitCode = 1
	if err := cache.Save(ctx, r2); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := cache.GetLatest(ctx, "alice/golden-pizza-api")
	if err != nil {
		t.Fatalf("GetLatest() error = %v", err)
	}
	if got.Commit != "bbb222" {
		t.Fatalf("latest commit = %s, want bbb222", got.Commit)
	}

	// First commit should still be available by its key
	old, err := cache.GetByCommit(ctx, "alice/golden-pizza-api", "aaa111")
	if err != nil {
		t.Fatalf("GetByCommit() error = %v", err)
	}
	if old.Commit != "aaa111" {
		t.Fatalf("old commit = %s, want aaa111", old.Commit)
	}
}

func TestReportCache_TTLExpiry(t *testing.T) {
	client := newTestRedis(t)
	flushRedis(t, client)

	cache := rediscache.NewReportCache(client, 100*time.Millisecond)
	ctx := context.Background()

	report := sampleReport("alice/golden-pizza-api", "ttl999", domain.StatusPassed)
	if err := cache.Save(ctx, report); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	_, err := cache.GetLatest(ctx, "alice/golden-pizza-api")
	if err != redis.Nil {
		t.Fatalf("error = %v, want redis.Nil after TTL expiry", err)
	}
}
