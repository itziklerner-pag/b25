package executor

import (
	"github.com/go-redis/redis/v8"
	"github.com/nats-io/nats.go"
)

// GetRedisClient returns the Redis client
func (e *OrderExecutor) GetRedisClient() *redis.Client {
	return e.redisClient
}

// GetNATSConn returns the NATS connection
func (e *OrderExecutor) GetNATSConn() *nats.Conn {
	return e.natsConn
}
