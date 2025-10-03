package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/b25/strategy-engine/internal/config"
	"github.com/b25/strategy-engine/internal/strategies"
	"github.com/b25/strategy-engine/pkg/logger"
	"github.com/b25/strategy-engine/pkg/metrics"
	"github.com/go-redis/redis/v8"
)

// RedisSubscriber handles Redis pub/sub for market data
type RedisSubscriber struct {
	client  *redis.Client
	logger  *logger.Logger
	metrics *metrics.Collector
	handler MarketDataHandler
}

// MarketDataHandler processes market data
type MarketDataHandler func(*strategies.MarketData) error

// NewRedisSubscriber creates a new Redis subscriber
func NewRedisSubscriber(cfg *config.RedisConfig, log *logger.Logger, m *metrics.Collector) (*RedisSubscriber, error) {
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

	log.Info("Connected to Redis",
		"host", cfg.Host,
		"port", cfg.Port,
	)

	return &RedisSubscriber{
		client:  client,
		logger:  log,
		metrics: m,
	}, nil
}

// Subscribe subscribes to market data channels
func (s *RedisSubscriber) Subscribe(ctx context.Context, channels []string, handler MarketDataHandler) error {
	s.handler = handler

	pubsub := s.client.Subscribe(ctx, channels...)
	defer pubsub.Close()

	// Wait for confirmation
	_, err := pubsub.Receive(ctx)
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	s.logger.Info("Subscribed to market data channels", "channels", channels)

	// Start receiving messages
	ch := pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Market data subscription stopped")
			return nil

		case msg := <-ch:
			if msg == nil {
				continue
			}

			startTime := time.Now()

			// Parse market data
			var marketData strategies.MarketData
			if err := json.Unmarshal([]byte(msg.Payload), &marketData); err != nil {
				s.logger.Error("Failed to parse market data",
					"error", err,
					"channel", msg.Channel,
				)
				continue
			}

			// Track latency (time since data was created)
			if !marketData.Timestamp.IsZero() {
				latency := time.Since(marketData.Timestamp).Microseconds()
				s.metrics.MarketDataLatency.WithLabelValues(marketData.Symbol).Observe(float64(latency))
			}

			// Handle market data
			if err := s.handler(&marketData); err != nil {
				s.logger.Error("Failed to handle market data",
					"error", err,
					"symbol", marketData.Symbol,
				)
				continue
			}

			// Track metrics
			s.metrics.MarketDataReceived.WithLabelValues(marketData.Symbol, marketData.Type).Inc()

			processingTime := time.Since(startTime).Microseconds()
			s.metrics.ProcessingTime.WithLabelValues("market_data").Observe(float64(processingTime))
		}
	}
}

// Publish publishes a message to a channel
func (s *RedisSubscriber) Publish(ctx context.Context, channel string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := s.client.Publish(ctx, channel, payload).Err(); err != nil {
		return fmt.Errorf("failed to publish: %w", err)
	}

	return nil
}

// Close closes the Redis connection
func (s *RedisSubscriber) Close() error {
	return s.client.Close()
}
