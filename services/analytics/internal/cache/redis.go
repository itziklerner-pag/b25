package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/b25/analytics/internal/config"
	"github.com/b25/analytics/internal/models"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// RedisCache handles caching operations
type RedisCache struct {
	client *redis.Client
	logger *zap.Logger
	ttl    time.Duration
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(cfg *config.RedisConfig, ttl time.Duration, logger *zap.Logger) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Connected to Redis",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.Int("db", cfg.DB),
	)

	return &RedisCache{
		client: client,
		logger: logger,
		ttl:    ttl,
	}, nil
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// GetClient returns the underlying Redis client for advanced operations
func (c *RedisCache) GetClient() *redis.Client {
	return c.client
}

// GetDashboardMetrics retrieves cached dashboard metrics
func (c *RedisCache) GetDashboardMetrics(ctx context.Context) (*models.DashboardMetrics, error) {
	key := "dashboard:metrics"

	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get from cache: %w", err)
	}

	var metrics models.DashboardMetrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metrics: %w", err)
	}

	return &metrics, nil
}

// SetDashboardMetrics caches dashboard metrics
func (c *RedisCache) SetDashboardMetrics(ctx context.Context, metrics *models.DashboardMetrics) error {
	key := "dashboard:metrics"

	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	if err := c.client.Set(ctx, key, data, c.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// GetQueryResult retrieves cached query result
func (c *RedisCache) GetQueryResult(ctx context.Context, cacheKey string) (*models.QueryResult, error) {
	data, err := c.client.Get(ctx, cacheKey).Bytes()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get from cache: %w", err)
	}

	var result models.QueryResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &result, nil
}

// SetQueryResult caches query result
func (c *RedisCache) SetQueryResult(ctx context.Context, cacheKey string, result *models.QueryResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	if err := c.client.Set(ctx, cacheKey, data, c.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// IncrementEventCounter increments an event counter
func (c *RedisCache) IncrementEventCounter(ctx context.Context, eventType string) error {
	key := fmt.Sprintf("counter:events:%s", eventType)
	return c.client.Incr(ctx, key).Err()
}

// GetEventCounter retrieves event counter value
func (c *RedisCache) GetEventCounter(ctx context.Context, eventType string) (int64, error) {
	key := fmt.Sprintf("counter:events:%s", eventType)
	return c.client.Get(ctx, key).Int64()
}

// SetActiveUsers sets active users count
func (c *RedisCache) SetActiveUsers(ctx context.Context, userID string, ttl time.Duration) error {
	key := "active:users"
	return c.client.SAdd(ctx, key, userID).Err()
}

// GetActiveUsersCount gets active users count
func (c *RedisCache) GetActiveUsersCount(ctx context.Context) (int64, error) {
	key := "active:users"
	return c.client.SCard(ctx, key).Result()
}

// InvalidateCache invalidates a cache key
func (c *RedisCache) InvalidateCache(ctx context.Context, pattern string) error {
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			c.logger.Error("Failed to delete cache key",
				zap.Error(err),
				zap.String("key", iter.Val()),
			)
		}
	}
	return iter.Err()
}

// GetHealth checks Redis health
func (c *RedisCache) GetHealth(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// GenerateCacheKey generates a cache key for a query
func GenerateCacheKey(prefix string, params map[string]interface{}) string {
	// Simple cache key generation - could be more sophisticated
	data, _ := json.Marshal(params)
	return fmt.Sprintf("%s:%s", prefix, string(data))
}
