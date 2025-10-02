package aggregation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/b25/analytics/internal/config"
	"github.com/b25/analytics/internal/models"
	"github.com/b25/analytics/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Engine handles metric aggregation
type Engine struct {
	cfg       *config.AggregationConfig
	repo      *repository.Repository
	logger    *zap.Logger
	intervals []Interval
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// Interval represents a time aggregation interval
type Interval struct {
	Name     string
	Duration time.Duration
}

// NewEngine creates a new aggregation engine
func NewEngine(cfg *config.AggregationConfig, repo *repository.Repository, logger *zap.Logger) (*Engine, error) {
	ctx, cancel := context.WithCancel(context.Background())

	intervals, err := parseIntervals(cfg.Intervals)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to parse intervals: %w", err)
	}

	return &Engine{
		cfg:       cfg,
		repo:      repo,
		logger:    logger,
		intervals: intervals,
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}

// Start starts the aggregation engine
func (e *Engine) Start() error {
	e.logger.Info("Starting aggregation engine",
		zap.Int("workers", e.cfg.Workers),
		zap.Int("intervals", len(e.intervals)),
	)

	// Start aggregation workers for each interval
	for _, interval := range e.intervals {
		e.wg.Add(1)
		go e.runAggregation(interval)
	}

	return nil
}

// Stop stops the aggregation engine
func (e *Engine) Stop() error {
	e.logger.Info("Stopping aggregation engine")
	e.cancel()
	e.wg.Wait()
	e.logger.Info("Aggregation engine stopped")
	return nil
}

// runAggregation runs periodic aggregation for a specific interval
func (e *Engine) runAggregation(interval Interval) {
	defer e.wg.Done()

	ticker := time.NewTicker(interval.Duration)
	defer ticker.Stop()

	e.logger.Info("Started aggregation worker",
		zap.String("interval", interval.Name),
		zap.Duration("duration", interval.Duration),
	)

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			e.performAggregation(interval)
		}
	}
}

// performAggregation performs aggregation for a specific interval
func (e *Engine) performAggregation(interval Interval) {
	start := time.Now()

	// Calculate time bucket
	now := time.Now().UTC()
	timeBucket := now.Truncate(interval.Duration)

	e.logger.Debug("Performing aggregation",
		zap.String("interval", interval.Name),
		zap.Time("time_bucket", timeBucket),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Aggregate event counts by type
	if err := e.aggregateEventCounts(ctx, interval, timeBucket); err != nil {
		e.logger.Error("Failed to aggregate event counts",
			zap.Error(err),
			zap.String("interval", interval.Name),
		)
	}

	// Aggregate trading metrics
	if err := e.aggregateTradingMetrics(ctx, interval, timeBucket); err != nil {
		e.logger.Error("Failed to aggregate trading metrics",
			zap.Error(err),
			zap.String("interval", interval.Name),
		)
	}

	duration := time.Since(start)
	e.logger.Info("Aggregation completed",
		zap.String("interval", interval.Name),
		zap.Duration("duration", duration),
		zap.Time("time_bucket", timeBucket),
	)
}

// aggregateEventCounts aggregates event counts by type
func (e *Engine) aggregateEventCounts(ctx context.Context, interval Interval, timeBucket time.Time) error {
	startTime := timeBucket
	endTime := timeBucket.Add(interval.Duration)

	counts, err := e.repo.GetEventCountByType(ctx, startTime, endTime)
	if err != nil {
		return fmt.Errorf("failed to get event counts: %w", err)
	}

	for eventType, count := range counts {
		metric := &models.MetricAggregation{
			ID:         uuid.New(),
			MetricName: fmt.Sprintf("events.count.%s", eventType),
			Interval:   interval.Name,
			TimeBucket: timeBucket,
			Count:      count,
			Dimensions: map[string]interface{}{
				"event_type": eventType,
			},
			CreatedAt: time.Now(),
		}

		if err := e.repo.InsertMetricAggregation(ctx, metric); err != nil {
			e.logger.Error("Failed to insert metric aggregation",
				zap.Error(err),
				zap.String("metric", metric.MetricName),
			)
		}
	}

	return nil
}

// aggregateTradingMetrics aggregates trading-specific metrics
func (e *Engine) aggregateTradingMetrics(ctx context.Context, interval Interval, timeBucket time.Time) error {
	// This would include:
	// - Order fill rates
	// - Average latency
	// - Volume metrics
	// - P&L aggregations
	// - Strategy performance

	// Placeholder for now - would be implemented with actual trading queries
	e.logger.Debug("Trading metrics aggregation placeholder",
		zap.String("interval", interval.Name),
		zap.Time("time_bucket", timeBucket),
	)

	return nil
}

// parseIntervals converts interval strings to Interval structs
func parseIntervals(intervalStrings []string) ([]Interval, error) {
	var intervals []Interval

	for _, str := range intervalStrings {
		duration, err := parseIntervalDuration(str)
		if err != nil {
			return nil, fmt.Errorf("invalid interval %s: %w", str, err)
		}

		intervals = append(intervals, Interval{
			Name:     str,
			Duration: duration,
		})
	}

	return intervals, nil
}

// parseIntervalDuration converts interval string to time.Duration
func parseIntervalDuration(interval string) (time.Duration, error) {
	switch interval {
	case "1m":
		return time.Minute, nil
	case "5m":
		return 5 * time.Minute, nil
	case "15m":
		return 15 * time.Minute, nil
	case "30m":
		return 30 * time.Minute, nil
	case "1h":
		return time.Hour, nil
	case "4h":
		return 4 * time.Hour, nil
	case "1d":
		return 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unsupported interval: %s", interval)
	}
}
