package e2e_test_helpers

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
)

func RateLimiterKeyExists(t *testing.T, rdb *redis.Client, key string) bool {
	t.Helper()
	exists, err := rdb.Exists(context.Background(), key).Result()
	if err != nil {
		t.Fatalf("redis exists %s: %v", key, err)
	}
	return exists > 0
}

func RateLimiterWindowCount(t *testing.T, rdb *redis.Client, windowKey string) int64 {
	t.Helper()
	count, err := rdb.Get(context.Background(), windowKey).Int64()
	if err != nil {
		t.Fatalf("redis get %s: %v", windowKey, err)
	}
	return count
}
