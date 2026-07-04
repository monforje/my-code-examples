// Package rediscache
package rediscache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"notifications/internal/models/domain"

	"github.com/redis/go-redis/v9"
)

// ReportCache — кеш CI-отчётов в Redis.
type ReportCache struct {
	client *redis.Client
	ttl    time.Duration
}

// NewReportCache - конструктор кеша отчётов.
func NewReportCache(client *redis.Client, ttl time.Duration) *ReportCache {
	return &ReportCache{
		client: client,
		ttl:    ttl,
	}
}

// Save - сохраняет отчёт в кеш по UID (последний) и по commit.
/*
    1. Сериализовать отчёт в JSON.
    2. Записать по ключу latest (uid).
    3. Записать по ключу commit (uid:commit).
    4. Установить TTL на оба ключа.
*/
func (c *ReportCache) Save(ctx context.Context, report *domain.CIReport) error {
	// 1.
	data, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}

	// 2.
	latestKey := latestKey(report.UID)
	if err := c.client.Set(ctx, latestKey, data, c.ttl).Err(); err != nil {
		return fmt.Errorf("set latest: %w", err)
	}

	// 3.
	commitKey := commitKey(report.UID, report.Commit)
	if err := c.client.Set(ctx, commitKey, data, c.ttl).Err(); err != nil {
		return fmt.Errorf("set commit: %w", err)
	}

	// 4.
	return nil
}

// GetLatest - возвращает последний отчёт по UID.
func (c *ReportCache) GetLatest(ctx context.Context, uid string) (*domain.CIReport, error) {
	data, err := c.client.Get(ctx, latestKey(uid)).Bytes()
	if err != nil {
		return nil, err
	}

	var report domain.CIReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("unmarshal report: %w", err)
	}
	return &report, nil
}

// GetByCommit - возвращает отчёт по UID и commit SHA.
func (c *ReportCache) GetByCommit(ctx context.Context, uid, commit string) (*domain.CIReport, error) {
	data, err := c.client.Get(ctx, commitKey(uid, commit)).Bytes()
	if err != nil {
		return nil, err
	}

	var report domain.CIReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("unmarshal report: %w", err)
	}
	return &report, nil
}

func latestKey(uid string) string {
	return fmt.Sprintf("ci:report:%s:latest", uid)
}

func commitKey(uid, commit string) string {
	return fmt.Sprintf("ci:report:%s:%s", uid, commit)
}
