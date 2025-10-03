package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// WebSocket connection metrics
	ConnectedClients = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dashboard_connected_clients",
			Help: "Number of connected WebSocket clients",
		},
		[]string{"client_type"},
	)

	MessagesSent = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dashboard_messages_sent_total",
			Help: "Total number of messages sent to clients",
		},
		[]string{"client_type", "message_type"},
	)

	MessagesReceived = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dashboard_messages_received_total",
			Help: "Total number of messages received from clients",
		},
		[]string{"message_type"},
	)

	BroadcastLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dashboard_broadcast_latency_seconds",
			Help:    "Broadcast latency in seconds",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 10),
		},
		[]string{"client_type"},
	)

	SerializationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dashboard_serialization_duration_seconds",
			Help:    "Serialization duration in seconds",
			Buckets: prometheus.ExponentialBuckets(0.0001, 2, 10),
		},
		[]string{"format"},
	)

	MessageSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dashboard_message_size_bytes",
			Help:    "Size of messages in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 2, 10),
		},
		[]string{"format", "message_type"},
	)

	StateUpdateLag = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dashboard_state_update_lag_seconds",
			Help: "Lag between state source update and cache update",
		},
		[]string{"source"},
	)

	ClientSubscriptions = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dashboard_client_subscriptions",
			Help: "Number of active subscriptions per channel",
		},
		[]string{"channel"},
	)

	ActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "dashboard_active_connections",
			Help: "Total number of active WebSocket connections",
		},
	)
)

// Helper functions to update metrics
func IncrementConnectedClients(clientType string) {
	ConnectedClients.WithLabelValues(clientType).Inc()
	ActiveConnections.Inc()
}

func DecrementConnectedClients(clientType string) {
	ConnectedClients.WithLabelValues(clientType).Dec()
	ActiveConnections.Dec()
}

func RecordMessageSent(clientType, messageType string) {
	MessagesSent.WithLabelValues(clientType, messageType).Inc()
}

func RecordMessageReceived(messageType string) {
	MessagesReceived.WithLabelValues(messageType).Inc()
}

func RecordBroadcastLatency(clientType string, duration float64) {
	BroadcastLatency.WithLabelValues(clientType).Observe(duration)
}

func RecordSerializationDuration(format string, duration float64) {
	SerializationDuration.WithLabelValues(format).Observe(duration)
}

func RecordMessageSize(format, messageType string, size int) {
	MessageSize.WithLabelValues(format, messageType).Observe(float64(size))
}

func RecordStateUpdateLag(source string, lag float64) {
	StateUpdateLag.WithLabelValues(source).Set(lag)
}

func IncrementClientSubscriptions(channel string) {
	ClientSubscriptions.WithLabelValues(channel).Inc()
}

func DecrementClientSubscriptions(channel string) {
	ClientSubscriptions.WithLabelValues(channel).Dec()
}
