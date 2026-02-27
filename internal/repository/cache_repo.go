package repository

import (
	"context"
	"encoding/json"
	"time"
	"url/internal/models"

	"github.com/redis/go-redis/v9"
)

// RedisCache provides a caching layer for short URLs using Redis.
// Improves performance by reducing database lookups
type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedisCache initializes Redis cache with TTL
func NewRedisCache(client *redis.Client, ttl time.Duration) *RedisCache {
	return &RedisCache{client: client, ttl: ttl}
}

// Set stores a URL in Redis with a TTL
func (r *RedisCache) Set(c context.Context, url models.URL) error {
	// Convert struct to JSON for storage
	data, err := json.Marshal(url)
	if err != nil {
		return err
	}
	// Store with expiration
	return r.client.Set(c, cacheKey(string(url.ShortCode)), data, r.ttl).Err()
}

// Get retrieves a URL from Redis
// Returns (nil, nil) if key does not exist — cache miss, not an error
func (r *RedisCache) Get(c context.Context, code string) (*models.URL, error) {
	data, err := r.client.Get(c, cacheKey(code)).Bytes()

	// Cache miss — tell the service to fall back to Postgres
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var url models.URL
	// Deserialize JSON into struct
	if err := json.Unmarshal(data, &url); err != nil {
		return nil, err
	}
	return &url, nil
}

// Delete removes a cached URL entry
func (r *RedisCache) Delete(c context.Context, code string) error {
	return r.client.Del(c, cacheKey(code)).Err()
}

// cacheKey creates a consistent Redis key format
func cacheKey(code string) string {
	return "url:" + code
}
