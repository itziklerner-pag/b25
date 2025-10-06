package ingestion

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/b25/analytics/internal/config"
	"github.com/b25/analytics/internal/metrics"
	"github.com/b25/analytics/internal/models"
	"github.com/b25/analytics/internal/repository"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// Consumer handles Kafka event consumption and ingestion
type Consumer struct {
	cfg              *config.KafkaConfig
	ingestionCfg     *config.IngestionConfig
	repo             *repository.Repository
	prometheusMetrics *metrics.Metrics
	logger           *zap.Logger
	readers          []*kafka.Reader
	eventChan        chan *models.Event
	wg               sync.WaitGroup
	ctx              context.Context
	cancel           context.CancelFunc
	metrics          *IngestionMetrics
}

// IngestionMetrics tracks ingestion performance
type IngestionMetrics struct {
	mu                sync.RWMutex
	EventsIngested    int64
	EventsFailed      int64
	BatchesProcessed  int64
	LastBatchSize     int
	LastBatchDuration time.Duration
	TotalLatency      time.Duration
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(cfg *config.KafkaConfig, ingestionCfg *config.IngestionConfig, repo *repository.Repository, promMetrics *metrics.Metrics, logger *zap.Logger) *Consumer {
	ctx, cancel := context.WithCancel(context.Background())

	c := &Consumer{
		cfg:              cfg,
		ingestionCfg:     ingestionCfg,
		repo:             repo,
		prometheusMetrics: promMetrics,
		logger:           logger,
		eventChan:        make(chan *models.Event, ingestionCfg.BufferSize),
		ctx:              ctx,
		cancel:           cancel,
		metrics:          &IngestionMetrics{},
	}

	// Create Kafka readers for each topic
	for _, topic := range cfg.Topics {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:        cfg.Brokers,
			GroupID:        cfg.ConsumerGroup,
			Topic:          topic,
			MinBytes:       10e3, // 10KB
			MaxBytes:       10e6, // 10MB
			CommitInterval: time.Second,
			StartOffset:    kafka.LastOffset,
			MaxWait:        500 * time.Millisecond,
		})
		c.readers = append(c.readers, reader)
	}

	return c
}

// Start starts the consumer
func (c *Consumer) Start() error {
	c.logger.Info("Starting event consumer",
		zap.Int("workers", c.ingestionCfg.Workers),
		zap.Int("buffer_size", c.ingestionCfg.BufferSize),
		zap.Int("batch_size", c.ingestionCfg.BatchSize),
	)

	// Start Kafka readers
	for _, reader := range c.readers {
		c.wg.Add(1)
		go c.consumeMessages(reader)
	}

	// Start batch processors
	for i := 0; i < c.ingestionCfg.Workers; i++ {
		c.wg.Add(1)
		go c.processBatches(i)
	}

	return nil
}

// Stop stops the consumer gracefully
func (c *Consumer) Stop() error {
	c.logger.Info("Stopping event consumer")
	c.cancel()

	// Close Kafka readers
	for _, reader := range c.readers {
		if err := reader.Close(); err != nil {
			c.logger.Error("Failed to close Kafka reader", zap.Error(err))
		}
	}

	// Wait for all goroutines to finish
	c.wg.Wait()

	// Close event channel
	close(c.eventChan)

	c.logger.Info("Event consumer stopped",
		zap.Int64("total_events_ingested", c.metrics.EventsIngested),
		zap.Int64("total_events_failed", c.metrics.EventsFailed),
		zap.Int64("total_batches", c.metrics.BatchesProcessed),
	)

	return nil
}

// consumeMessages reads messages from Kafka and sends them to the event channel
func (c *Consumer) consumeMessages(reader *kafka.Reader) {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			msg, err := reader.FetchMessage(c.ctx)
			if err != nil {
				if err == context.Canceled {
					return
				}
				c.logger.Error("Failed to fetch message", zap.Error(err))
				continue
			}

			event, err := c.parseEvent(msg.Value)
			if err != nil {
				c.logger.Error("Failed to parse event",
					zap.Error(err),
					zap.String("topic", msg.Topic),
					zap.Int("partition", msg.Partition),
					zap.Int64("offset", msg.Offset),
				)
				c.incrementFailedEvents()
				continue
			}

			// Send event to processing channel
			select {
			case c.eventChan <- event:
			case <-c.ctx.Done():
				return
			}

			// Commit the message
			if err := reader.CommitMessages(c.ctx, msg); err != nil {
				c.logger.Error("Failed to commit message", zap.Error(err))
			}
		}
	}
}

// processBatches processes events in batches for efficient database insertion
func (c *Consumer) processBatches(workerID int) {
	defer c.wg.Done()

	batch := make([]*models.Event, 0, c.ingestionCfg.BatchSize)
	ticker := time.NewTicker(c.ingestionCfg.BatchTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			// Flush remaining events
			if len(batch) > 0 {
				c.flushBatch(batch, workerID)
			}
			return

		case event := <-c.eventChan:
			batch = append(batch, event)

			// Flush batch if it reaches the configured size
			if len(batch) >= c.ingestionCfg.BatchSize {
				c.flushBatch(batch, workerID)
				batch = make([]*models.Event, 0, c.ingestionCfg.BatchSize)
				ticker.Reset(c.ingestionCfg.BatchTimeout)
			}

		case <-ticker.C:
			// Flush batch on timeout
			if len(batch) > 0 {
				c.flushBatch(batch, workerID)
				batch = make([]*models.Event, 0, c.ingestionCfg.BatchSize)
			}
		}
	}
}

// flushBatch inserts a batch of events into the database
func (c *Consumer) flushBatch(batch []*models.Event, workerID int) {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := c.repo.InsertEventsBatch(ctx, batch)
	duration := time.Since(start)

	if err != nil {
		c.logger.Error("Failed to insert batch",
			zap.Error(err),
			zap.Int("worker_id", workerID),
			zap.Int("batch_size", len(batch)),
			zap.Duration("duration", duration),
		)
		c.metrics.mu.Lock()
		c.metrics.EventsFailed += int64(len(batch))
		c.metrics.mu.Unlock()

		// Update Prometheus metrics
		c.prometheusMetrics.EventsFailed.Add(float64(len(batch)))
		return
	}

	// Update metrics
	c.metrics.mu.Lock()
	c.metrics.EventsIngested += int64(len(batch))
	c.metrics.BatchesProcessed++
	c.metrics.LastBatchSize = len(batch)
	c.metrics.LastBatchDuration = duration
	c.metrics.TotalLatency += duration
	c.metrics.mu.Unlock()

	// Update Prometheus metrics
	c.prometheusMetrics.EventsIngested.Add(float64(len(batch)))
	c.prometheusMetrics.BatchesProcessed.Inc()
	c.prometheusMetrics.BatchDuration.Observe(duration.Seconds())

	c.logger.Debug("Batch inserted successfully",
		zap.Int("worker_id", workerID),
		zap.Int("batch_size", len(batch)),
		zap.Duration("duration", duration),
		zap.Float64("events_per_sec", float64(len(batch))/duration.Seconds()),
	)
}

// parseEvent parses a Kafka message into an Event model
func (c *Consumer) parseEvent(data []byte) (*models.Event, error) {
	var rawEvent map[string]interface{}
	if err := json.Unmarshal(data, &rawEvent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event: %w", err)
	}

	event := &models.Event{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
	}

	// Parse event type
	if eventType, ok := rawEvent["event_type"].(string); ok {
		event.EventType = eventType
	} else {
		return nil, fmt.Errorf("missing event_type field")
	}

	// Parse optional user ID
	if userID, ok := rawEvent["user_id"].(string); ok {
		event.UserID = &userID
	}

	// Parse optional session ID
	if sessionID, ok := rawEvent["session_id"].(string); ok {
		event.SessionID = &sessionID
	}

	// Parse timestamp
	if timestamp, ok := rawEvent["timestamp"].(string); ok {
		t, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			return nil, fmt.Errorf("invalid timestamp format: %w", err)
		}
		event.Timestamp = t
	} else {
		event.Timestamp = time.Now()
	}

	// Parse properties
	if properties, ok := rawEvent["properties"].(map[string]interface{}); ok {
		event.Properties = properties
	} else {
		event.Properties = make(map[string]interface{})
	}

	return event, nil
}

// incrementFailedEvents safely increments the failed events counter
func (c *Consumer) incrementFailedEvents() {
	c.metrics.mu.Lock()
	c.metrics.EventsFailed++
	c.metrics.mu.Unlock()
}

// GetMetrics returns current ingestion metrics
func (c *Consumer) GetMetrics() *IngestionMetrics {
	c.metrics.mu.RLock()
	defer c.metrics.mu.RUnlock()

	return &IngestionMetrics{
		EventsIngested:    c.metrics.EventsIngested,
		EventsFailed:      c.metrics.EventsFailed,
		BatchesProcessed:  c.metrics.BatchesProcessed,
		LastBatchSize:     c.metrics.LastBatchSize,
		LastBatchDuration: c.metrics.LastBatchDuration,
		TotalLatency:      c.metrics.TotalLatency,
	}
}
