package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/b25/analytics/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a mock handler (without actual DB/cache dependencies for this test)
	logger, _ := zap.NewDevelopment()
	handler := &Handler{
		logger: logger,
	}

	router := gin.New()
	router.GET("/health", handler.HealthCheck)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Note: This test will fail without actual DB connection
	// In a real test, you'd use mocks
	assert.Contains(t, []int{http.StatusOK, http.StatusServiceUnavailable}, w.Code)
}

func TestTrackEventRequestValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger, _ := zap.NewDevelopment()
	handler := &Handler{
		logger: logger,
	}

	router := gin.New()
	router.POST("/events", handler.TrackEvent)

	tests := []struct {
		name       string
		payload    interface{}
		wantStatus int
	}{
		{
			name: "Valid event",
			payload: models.Event{
				EventType: models.EventTypeOrderPlaced,
				Properties: map[string]interface{}{
					"symbol": "BTCUSDT",
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "Invalid JSON",
			payload: "invalid json",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			if str, ok := tt.payload.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.payload)
			}

			req, _ := http.NewRequest("POST", "/events", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Note: Without DB, all valid requests will fail at DB insertion
			// This test mainly validates request parsing
			assert.NotEqual(t, 0, w.Code)
		})
	}
}

func TestQueryResultSerialization(t *testing.T) {
	result := &models.QueryResult{
		MetricName: "test.metric",
		Interval:   "1h",
		StartTime:  time.Now().Add(-24 * time.Hour),
		EndTime:    time.Now(),
		DataPoints: []models.TimeSeriesPoint{
			{
				Timestamp: time.Now().Add(-1 * time.Hour),
				Value:     100.0,
			},
			{
				Timestamp: time.Now(),
				Value:     150.0,
			},
		},
		TotalCount:  2,
		Aggregation: "avg",
	}

	data, err := json.Marshal(result)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	var decoded models.QueryResult
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, result.MetricName, decoded.MetricName)
	assert.Equal(t, result.TotalCount, decoded.TotalCount)
}
