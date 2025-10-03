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
	"github.com/nats-io/nats.go"
)

// NATSSubscriber handles NATS pub/sub for fills and positions
type NATSSubscriber struct {
	conn         *nats.Conn
	logger       *logger.Logger
	metrics      *metrics.Collector
	fillHandler  FillHandler
	posHandler   PositionHandler
}

// FillHandler processes fill events
type FillHandler func(*strategies.Fill) error

// PositionHandler processes position updates
type PositionHandler func(*strategies.Position) error

// NewNATSSubscriber creates a new NATS subscriber
func NewNATSSubscriber(cfg *config.NATSConfig, log *logger.Logger, m *metrics.Collector) (*NATSSubscriber, error) {
	opts := []nats.Option{
		nats.MaxReconnects(cfg.MaxReconnects),
		nats.ReconnectWait(time.Duration(cfg.ReconnectWait) * time.Second),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				log.Warn("NATS disconnected", "error", err)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Info("NATS reconnected", "url", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Info("NATS connection closed")
		}),
	}

	conn, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	log.Info("Connected to NATS", "url", cfg.URL)

	return &NATSSubscriber{
		conn:    conn,
		logger:  log,
		metrics: m,
	}, nil
}

// SubscribeFills subscribes to fill events
func (s *NATSSubscriber) SubscribeFills(ctx context.Context, subject string, handler FillHandler) error {
	s.fillHandler = handler

	sub, err := s.conn.Subscribe(subject, func(msg *nats.Msg) {
		startTime := time.Now()

		// Parse fill
		var fill strategies.Fill
		if err := json.Unmarshal(msg.Data, &fill); err != nil {
			s.logger.Error("Failed to parse fill",
				"error", err,
				"subject", msg.Subject,
			)
			return
		}

		// Track latency
		if !fill.Timestamp.IsZero() {
			latency := time.Since(fill.Timestamp).Microseconds()
			s.metrics.FillLatency.WithLabelValues(fill.Strategy).Observe(float64(latency))
		}

		// Handle fill
		if err := s.fillHandler(&fill); err != nil {
			s.logger.Error("Failed to handle fill",
				"error", err,
				"fill_id", fill.FillID,
			)
			return
		}

		// Track metrics
		s.metrics.FillsReceived.WithLabelValues(fill.Strategy, fill.Symbol, fill.Side).Inc()

		processingTime := time.Since(startTime).Microseconds()
		s.metrics.ProcessingTime.WithLabelValues("fill").Observe(float64(processingTime))
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to fills: %w", err)
	}

	s.logger.Info("Subscribed to fills", "subject", subject)

	// Wait for context cancellation
	<-ctx.Done()
	sub.Unsubscribe()

	return nil
}

// SubscribePositions subscribes to position updates
func (s *NATSSubscriber) SubscribePositions(ctx context.Context, subject string, handler PositionHandler) error {
	s.posHandler = handler

	sub, err := s.conn.Subscribe(subject, func(msg *nats.Msg) {
		startTime := time.Now()

		// Parse position
		var position strategies.Position
		if err := json.Unmarshal(msg.Data, &position); err != nil {
			s.logger.Error("Failed to parse position",
				"error", err,
				"subject", msg.Subject,
			)
			return
		}

		// Handle position
		if err := s.posHandler(&position); err != nil {
			s.logger.Error("Failed to handle position",
				"error", err,
				"symbol", position.Symbol,
			)
			return
		}

		// Track metrics
		s.metrics.PositionUpdates.WithLabelValues(position.Strategy, position.Symbol).Inc()
		s.metrics.CurrentPositions.WithLabelValues(position.Strategy, position.Symbol).Set(position.Quantity)

		processingTime := time.Since(startTime).Microseconds()
		s.metrics.ProcessingTime.WithLabelValues("position").Observe(float64(processingTime))
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to positions: %w", err)
	}

	s.logger.Info("Subscribed to positions", "subject", subject)

	// Wait for context cancellation
	<-ctx.Done()
	sub.Unsubscribe()

	return nil
}

// Publish publishes a message to a subject
func (s *NATSSubscriber) Publish(subject string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := s.conn.Publish(subject, payload); err != nil {
		return fmt.Errorf("failed to publish: %w", err)
	}

	return nil
}

// Close closes the NATS connection
func (s *NATSSubscriber) Close() error {
	s.conn.Close()
	return nil
}
