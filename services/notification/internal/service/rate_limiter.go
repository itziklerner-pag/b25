package service

import (
	"context"
	"fmt"
	"time"

	"github.com/b25/services/notification/internal/config"
	"github.com/b25/services/notification/internal/models"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// RateLimiter interface for rate limiting
type RateLimiter interface {
	CheckLimit(ctx context.Context, userID uuid.UUID, channel models.NotificationChannel) error
	IncrementCount(ctx context.Context, userID uuid.UUID, channel models.NotificationChannel) error
	GetCurrentCount(ctx context.Context, userID uuid.UUID, channel models.NotificationChannel) (int, error)
	ResetLimit(ctx context.Context, userID uuid.UUID, channel models.NotificationChannel) error
}

// RedisRateLimiter implements rate limiting using Redis
type RedisRateLimiter struct {
	redis *redis.Client
	cfg   *config.RateLimitConfig
}

// NewRedisRateLimiter creates a new Redis-based rate limiter
func NewRedisRateLimiter(redisClient *redis.Client, cfg *config.RateLimitConfig) RateLimiter {
	return &RedisRateLimiter{
		redis: redisClient,
		cfg:   cfg,
	}
}

// CheckLimit checks if the user has exceeded the rate limit for the channel
func (r *RedisRateLimiter) CheckLimit(ctx context.Context, userID uuid.UUID, channel models.NotificationChannel) error {
	key := r.getRateLimitKey(userID, channel)
	limit := r.getLimit(channel)

	count, err := r.redis.Get(ctx, key).Int()
	if err == redis.Nil {
		// Key doesn't exist, under limit
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get rate limit: %w", err)
	}

	if count >= limit {
		return fmt.Errorf("rate limit exceeded: %d/%d per hour", count, limit)
	}

	return nil
}

// IncrementCount increments the counter for the user and channel
func (r *RedisRateLimiter) IncrementCount(ctx context.Context, userID uuid.UUID, channel models.NotificationChannel) error {
	key := r.getRateLimitKey(userID, channel)

	pipe := r.redis.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, time.Hour)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to increment rate limit: %w", err)
	}

	// Check if we just hit the limit
	if incr.Val() == 1 {
		// First request in this window, set expiry
		r.redis.Expire(ctx, key, time.Hour)
	}

	return nil
}

// GetCurrentCount returns the current count for the user and channel
func (r *RedisRateLimiter) GetCurrentCount(ctx context.Context, userID uuid.UUID, channel models.NotificationChannel) (int, error) {
	key := r.getRateLimitKey(userID, channel)

	count, err := r.redis.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get current count: %w", err)
	}

	return count, nil
}

// ResetLimit resets the rate limit for a user and channel
func (r *RedisRateLimiter) ResetLimit(ctx context.Context, userID uuid.UUID, channel models.NotificationChannel) error {
	key := r.getRateLimitKey(userID, channel)
	return r.redis.Del(ctx, key).Err()
}

func (r *RedisRateLimiter) getRateLimitKey(userID uuid.UUID, channel models.NotificationChannel) string {
	// Key format: ratelimit:{user_id}:{channel}:{hour}
	hour := time.Now().Format("2006010215")
	return fmt.Sprintf("ratelimit:%s:%s:%s", userID.String(), channel, hour)
}

func (r *RedisRateLimiter) getLimit(channel models.NotificationChannel) int {
	switch channel {
	case models.ChannelEmail:
		return r.cfg.EmailPerHour
	case models.ChannelSMS:
		return r.cfg.SMSPerHour
	case models.ChannelPush:
		return r.cfg.PushPerHour
	default:
		return 100 // default limit
	}
}
