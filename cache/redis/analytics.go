package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"tx-processor/models"

	"github.com/redis/go-redis/v9"
)

type RedisAnalyticsCache struct {
	client      *redis.Client
	PrefixState string
	defaultTTL  time.Duration
}

func NewRedisAnalyticsCache(client *redis.Client) *RedisAnalyticsCache {
	return &RedisAnalyticsCache{
		client:      client,
		PrefixState: "analytics:",
		defaultTTL:  5 * time.Hour,
	}
}

func (r *RedisAnalyticsCache) buildKeyState(state string) string {
	return fmt.Sprintf("%s:%s", r.PrefixState, state)
}

func (r *RedisAnalyticsCache) Get(ctx context.Context, userID string) (*models.UserAnalytics, error) {
	key := r.buildKeyState(userID)

	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // Cache miss
	} else if err != nil {
		return nil, fmt.Errorf("failed to get user info from cache: %w", err)
	}

	var analytics models.UserAnalytics
	if err := json.Unmarshal([]byte(val), &analytics); err != nil {
		return nil, err
	}

	return &analytics, nil
}

func (r *RedisAnalyticsCache) Set(ctx context.Context, analytics models.UserAnalytics) error {
	key := r.buildKeyState(analytics.UserID)
	expiration := r.defaultTTL

	data, err := json.Marshal(analytics)
	if err != nil {
		return fmt.Errorf("failed to marshal user info: %w", err)
	}

	if err := r.client.Set(ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set user info in cache: %w", err)
	}

	return nil
}

func (r *RedisAnalyticsCache) Delete(ctx context.Context, userID string) error {
	key := r.buildKeyState(userID)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete user info in cache: %w", err)
	}

	return nil
}
