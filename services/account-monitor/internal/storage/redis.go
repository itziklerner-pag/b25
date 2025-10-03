package storage

import (
	"fmt"

	"github.com/go-redis/redis/v8"

	"github.com/yourorg/b25/services/account-monitor/internal/config"
)

// NewRedisClient creates a new Redis client
func NewRedisClient(cfg config.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
}
