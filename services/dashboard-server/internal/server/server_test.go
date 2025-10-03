package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/yourusername/b25/services/dashboard-server/internal/aggregator"
	"github.com/yourusername/b25/services/dashboard-server/internal/broadcaster"
)

func TestNewServer(t *testing.T) {
	logger := zerolog.Nop()
	agg := aggregator.NewAggregator(logger, "localhost:6379")
	bcast := broadcaster.NewBroadcaster(logger, agg)

	server := NewServer(logger, agg, bcast)

	assert.NotNil(t, server)
	assert.NotNil(t, server.clients)
	assert.NotNil(t, server.aggregator)
	assert.NotNil(t, server.broadcaster)
}

func TestGenerateClientID(t *testing.T) {
	id1 := generateClientID()
	time.Sleep(time.Millisecond)
	id2 := generateClientID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
}

func TestHandleHistory(t *testing.T) {
	logger := zerolog.Nop()
	agg := aggregator.NewAggregator(logger, "localhost:6379")
	bcast := broadcaster.NewBroadcaster(logger, agg)
	server := NewServer(logger, agg, bcast)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "valid request",
			queryParams:    "?type=market_data&symbol=BTCUSDT&limit=100",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing type parameter",
			queryParams:    "?symbol=BTCUSDT",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/history"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			server.HandleHistory(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
