package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	urlIDSequenceKey = "url_id_sequence"
	urlKeyPrefix     = "url:"
)

type RedisURLCacheRepo struct {
	client *redis.Client
}

func NewRedisURLCacheRepo(client *redis.Client) *RedisURLCacheRepo {
	return &RedisURLCacheRepo{client: client}
}

// Create generates a new ID using INCR and stores the URL using pipeline for atomicity.
func (r *RedisURLCacheRepo) Create(ctx context.Context, longURL string, ttl time.Duration) (int64, error) {
	// First, get the ID using INCR
	id, err := r.client.Incr(ctx, urlIDSequenceKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to generate ID: %w", err)
	}

	// Use pipeline to SET the URL with TTL atomically
	key := fmt.Sprintf("%s%d", urlKeyPrefix, id)
	pipe := r.client.Pipeline()
	pipe.Set(ctx, key, longURL, ttl)

	_, err = pipe.Exec(ctx)
	if err != nil {
		// If SET fails, we've already incremented the counter
		// This creates a gap in the sequence, but that's acceptable
		return 0, fmt.Errorf("failed to set URL in cache: %w", err)
	}

	return id, nil
}

func (r *RedisURLCacheRepo) Set(ctx context.Context, id int64, longURL string, ttl time.Duration) error {
	key := fmt.Sprintf("%s%d", urlKeyPrefix, id)
	err := r.client.Set(ctx, key, longURL, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set URL in cache: %w", err)
	}
	return nil
}

func (r *RedisURLCacheRepo) Get(ctx context.Context, id int64) (string, error) {
	key := fmt.Sprintf("%s%d", urlKeyPrefix, id)
	longURL, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("URL not found or expired")
	}
	if err != nil {
		return "", fmt.Errorf("failed to get URL from cache: %w", err)
	}
	return longURL, nil
}

func (r *RedisURLCacheRepo) Delete(ctx context.Context, id int64) error {
	key := fmt.Sprintf("%s%d", urlKeyPrefix, id)
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete URL from cache: %w", err)
	}
	return nil
}
