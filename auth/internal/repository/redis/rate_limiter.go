package redisrepo

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrRateLimited = errors.New("rate limited")

type redisClient interface {
	SetNX(ctx context.Context, key string, value any, expiration time.Duration) *redis.BoolCmd
	Incr(ctx context.Context, key string) *redis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
}

type RateLimiter struct {
	client redisClient
}

func NewRateLimiter(client redisClient) *RateLimiter {
	return &RateLimiter{client: client}
}

func (r *RateLimiter) Allow(ctx context.Context, key string, cooldown, window time.Duration, maxRequests int64) error {
	cooldownKey := key + ":cooldown"
	ok, err := r.client.SetNX(ctx, cooldownKey, "1", cooldown).Result()
	if err != nil {
		return err
	}
	if !ok {
		return ErrRateLimited
	}

	windowKey := key + ":window"
	count, err := r.client.Incr(ctx, windowKey).Result()
	if err != nil {
		return err
	}

	if count == 1 {
		if err := r.client.Expire(ctx, windowKey, window).Err(); err != nil {
			return err
		}
	}

	if count > maxRequests {
		return ErrRateLimited
	}

	return nil
}
