package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/b25/api-gateway/internal/config"
	"github.com/b25/api-gateway/pkg/logger"
	"github.com/b25/api-gateway/pkg/metrics"
	"github.com/go-redis/redis/v8"
)

// Cache handles response caching
type Cache struct {
	client  *redis.Client
	config  config.CacheConfig
	log     *logger.Logger
	metrics *metrics.Collector
}

// NewCache creates a new cache instance
func NewCache(cfg config.CacheConfig, log *logger.Logger, m *metrics.Collector) (*Cache, error) {
	if !cfg.Enabled {
		return &Cache{
			config:  cfg,
			log:     log,
			metrics: m,
		}, nil
	}

	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Info("Connected to Redis cache", "url", cfg.RedisURL)

	return &Cache{
		client:  client,
		config:  cfg,
		log:     log,
		metrics: m,
	}, nil
}

// Get retrieves a value from cache
func (c *Cache) Get(ctx context.Context, key string) ([]byte, error) {
	if !c.config.Enabled || c.client == nil {
		return nil, redis.Nil
	}

	val, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			c.metrics.RecordCacheMiss(key)
			return nil, err
		}
		c.log.Errorw("Cache get error", "key", key, "error", err)
		return nil, err
	}

	c.metrics.RecordCacheHit(key)
	return val, nil
}

// Set stores a value in cache
func (c *Cache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if !c.config.Enabled || c.client == nil {
		return nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if ttl == 0 {
		ttl = c.config.DefaultTTL
	}

	err = c.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		c.log.Errorw("Cache set error", "key", key, "error", err)
		return err
	}

	return nil
}

// Delete removes a value from cache
func (c *Cache) Delete(ctx context.Context, keys ...string) error {
	if !c.config.Enabled || c.client == nil {
		return nil
	}

	err := c.client.Del(ctx, keys...).Err()
	if err != nil {
		c.log.Errorw("Cache delete error", "keys", keys, "error", err)
		return err
	}

	return nil
}

// Exists checks if a key exists in cache
func (c *Cache) Exists(ctx context.Context, key string) (bool, error) {
	if !c.config.Enabled || c.client == nil {
		return false, nil
	}

	count, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		c.log.Errorw("Cache exists error", "key", key, "error", err)
		return false, err
	}

	return count > 0, nil
}

// GetTTL returns the TTL for a given endpoint
func (c *Cache) GetTTL(endpoint string) time.Duration {
	if rule, exists := c.config.Rules[endpoint]; exists {
		return rule.TTL
	}
	return c.config.DefaultTTL
}

// GenerateKey generates a cache key from request parameters
func (c *Cache) GenerateKey(method, path, query string) string {
	if query != "" {
		return fmt.Sprintf("cache:%s:%s?%s", method, path, query)
	}
	return fmt.Sprintf("cache:%s:%s", method, path)
}

// Close closes the cache connection
func (c *Cache) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// InvalidatePattern invalidates all keys matching a pattern
func (c *Cache) InvalidatePattern(ctx context.Context, pattern string) error {
	if !c.config.Enabled || c.client == nil {
		return nil
	}

	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	keys := make([]string, 0)

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		c.log.Errorw("Cache scan error", "pattern", pattern, "error", err)
		return err
	}

	if len(keys) > 0 {
		return c.Delete(ctx, keys...)
	}

	return nil
}

// GetJSON retrieves and unmarshals a JSON value from cache
func (c *Cache) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := c.Get(ctx, key)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}

// SetJSON marshals and stores a JSON value in cache
func (c *Cache) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.Set(ctx, key, value, ttl)
}
