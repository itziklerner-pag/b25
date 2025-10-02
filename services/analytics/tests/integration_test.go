// +build integration

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/b25/analytics/internal/config"
	"github.com/b25/analytics/internal/models"
	"github.com/b25/analytics/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests require a running PostgreSQL instance
// Run with: go test -tags=integration ./tests/...

func setupTestDB(t *testing.T) *repository.Repository {
	cfg := &config.DatabaseConfig{
		Host:               "localhost",
		Port:               5432,
		Database:           "analytics_test",
		User:               "analytics_user",
		Password:           "analytics_password",
		SSLMode:            "disable",
		MaxConnections:     10,
		MaxIdleConnections: 5,
		ConnectionLifetime: 300 * time.Second,
	}

	repo, err := repository.NewRepository(cfg)
	require.NoError(t, err, "Failed to create repository")

	return repo
}

func TestEventInsertion(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	ctx := context.Background()

	event := &models.Event{
		ID:        uuid.New(),
		EventType: models.EventTypeOrderPlaced,
		UserID:    stringPtr("test-user"),
		SessionID: stringPtr("test-session"),
		Properties: map[string]interface{}{
			"symbol": "BTCUSDT",
			"price":  50000.0,
		},
		Timestamp: time.Now(),
		CreatedAt: time.Now(),
	}

	err := repo.InsertEvent(ctx, event)
	assert.NoError(t, err, "Failed to insert event")
}

func TestBatchEventInsertion(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	ctx := context.Background()

	events := make([]*models.Event, 100)
	for i := 0; i < 100; i++ {
		events[i] = &models.Event{
			ID:        uuid.New(),
			EventType: models.EventTypeOrderPlaced,
			UserID:    stringPtr("test-user"),
			SessionID: stringPtr("test-session"),
			Properties: map[string]interface{}{
				"symbol": "BTCUSDT",
				"index":  i,
			},
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
			CreatedAt: time.Now(),
		}
	}

	start := time.Now()
	err := repo.InsertEventsBatch(ctx, events)
	duration := time.Since(start)

	assert.NoError(t, err, "Failed to insert batch")
	t.Logf("Inserted %d events in %v (%.2f events/sec)", len(events), duration, float64(len(events))/duration.Seconds())
}

func TestEventQuery(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	ctx := context.Background()

	// Insert test events
	now := time.Now()
	for i := 0; i < 10; i++ {
		event := &models.Event{
			ID:        uuid.New(),
			EventType: models.EventTypeOrderPlaced,
			UserID:    stringPtr("query-test-user"),
			Timestamp: now.Add(time.Duration(i) * time.Minute),
			CreatedAt: now,
			Properties: make(map[string]interface{}),
		}
		err := repo.InsertEvent(ctx, event)
		require.NoError(t, err)
	}

	// Query events
	startTime := now.Add(-1 * time.Hour)
	endTime := now.Add(1 * time.Hour)

	events, err := repo.GetEventsByTimeRange(ctx, startTime, endTime, models.EventTypeOrderPlaced, 100)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(events), 10, "Should retrieve at least 10 events")
}

func TestMetricAggregation(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	ctx := context.Background()

	metric := &models.MetricAggregation{
		ID:         uuid.New(),
		MetricName: "test.metric",
		Interval:   "1h",
		TimeBucket: time.Now().Truncate(time.Hour),
		Count:      100,
		Sum:        float64Ptr(5000.0),
		Avg:        float64Ptr(50.0),
		Min:        float64Ptr(10.0),
		Max:        float64Ptr(100.0),
		Dimensions: map[string]interface{}{
			"test": "value",
		},
		CreatedAt: time.Now(),
	}

	err := repo.InsertMetricAggregation(ctx, metric)
	assert.NoError(t, err)

	// Query metric
	metrics, err := repo.GetMetricAggregations(
		ctx,
		"test.metric",
		"1h",
		metric.TimeBucket.Add(-1*time.Hour),
		metric.TimeBucket.Add(1*time.Hour),
	)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(metrics), 1)
}

func TestDashboardMetrics(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	ctx := context.Background()

	metrics, err := repo.GetDashboardMetrics(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.GreaterOrEqual(t, metrics.ActiveUsers, int64(0))
}

func TestCustomEventDefinition(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	ctx := context.Background()

	def := &models.CustomEventDefinition{
		ID:          uuid.New(),
		Name:        "test.custom.event",
		DisplayName: "Test Custom Event",
		Description: "A test custom event",
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"field1": map[string]string{"type": "string"},
			},
		},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.CreateCustomEventDefinition(ctx, def)
	assert.NoError(t, err)

	// Retrieve definition
	retrieved, err := repo.GetCustomEventDefinition(ctx, "test.custom.event")
	assert.NoError(t, err)
	assert.Equal(t, def.Name, retrieved.Name)
	assert.Equal(t, def.DisplayName, retrieved.DisplayName)
}

func TestHealthCheck(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	ctx := context.Background()

	err := repo.GetHealth(ctx)
	assert.NoError(t, err, "Health check should pass")
}

func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}
