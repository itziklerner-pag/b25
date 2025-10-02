package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestEventModel(t *testing.T) {
	event := &Event{
		ID:        uuid.New(),
		EventType: EventTypeOrderPlaced,
		UserID:    stringPtr("user123"),
		SessionID: stringPtr("session456"),
		Properties: map[string]interface{}{
			"symbol": "BTCUSDT",
			"price":  50000.0,
		},
		Timestamp: time.Now(),
		CreatedAt: time.Now(),
	}

	assert.NotEqual(t, uuid.Nil, event.ID)
	assert.Equal(t, EventTypeOrderPlaced, event.EventType)
	assert.Equal(t, "user123", *event.UserID)
	assert.NotNil(t, event.Properties)
	assert.Equal(t, "BTCUSDT", event.Properties["symbol"])
}

func TestEventTypeConstants(t *testing.T) {
	eventTypes := []string{
		EventTypeOrderPlaced,
		EventTypeOrderFilled,
		EventTypeOrderCanceled,
		EventTypeOrderRejected,
		EventTypeStrategyStarted,
		EventTypeStrategyStopped,
		EventTypeSignalGenerated,
		EventTypePositionOpened,
		EventTypePositionClosed,
		EventTypeBalanceUpdated,
	}

	for _, eventType := range eventTypes {
		assert.NotEmpty(t, eventType)
	}
}

func TestMetricAggregation(t *testing.T) {
	metric := &MetricAggregation{
		ID:         uuid.New(),
		MetricName: "events.count.order.placed",
		Interval:   "1m",
		TimeBucket: time.Now().Truncate(time.Minute),
		Count:      100,
		Sum:        float64Ptr(5000.0),
		Avg:        float64Ptr(50.0),
		Min:        float64Ptr(10.0),
		Max:        float64Ptr(100.0),
		Dimensions: map[string]interface{}{
			"event_type": "order.placed",
		},
		CreatedAt: time.Now(),
	}

	assert.NotEqual(t, uuid.Nil, metric.ID)
	assert.Equal(t, int64(100), metric.Count)
	assert.Equal(t, 50.0, *metric.Avg)
}

func TestDashboardMetrics(t *testing.T) {
	metrics := &DashboardMetrics{
		ActiveUsers:      100,
		EventsPerSecond:  50.5,
		OrdersPlaced:     1000,
		OrdersFilled:     950,
		ActiveStrategies: 5,
		TotalVolume:      1000000.0,
		AverageLatency:   10.5,
		ErrorRate:        0.01,
		CustomMetrics: map[string]float64{
			"custom_metric_1": 123.45,
		},
		LastUpdated: time.Now(),
	}

	assert.Equal(t, int64(100), metrics.ActiveUsers)
	assert.Equal(t, 50.5, metrics.EventsPerSecond)
	assert.NotNil(t, metrics.CustomMetrics)
}

func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}
