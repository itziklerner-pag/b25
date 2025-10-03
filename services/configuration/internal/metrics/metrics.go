package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// ConfigOperations tracks configuration operations
	ConfigOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "config_operations_total",
			Help: "Total number of configuration operations",
		},
		[]string{"operation", "type", "status"},
	)

	// ConfigOperationDuration tracks operation duration
	ConfigOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "config_operation_duration_seconds",
			Help:    "Duration of configuration operations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "type"},
	)

	// ActiveConfigurations tracks active configurations
	ActiveConfigurations = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "active_configurations",
			Help: "Number of active configurations by type",
		},
		[]string{"type"},
	)

	// ConfigVersions tracks configuration versions
	ConfigVersions = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "config_versions",
			Help: "Current version number of configurations",
		},
		[]string{"key"},
	)

	// ConfigUpdateEvents tracks NATS update events
	ConfigUpdateEvents = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "config_update_events_total",
			Help: "Total number of configuration update events published",
		},
		[]string{"type", "action"},
	)

	// ValidationErrors tracks validation errors
	ValidationErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "config_validation_errors_total",
			Help: "Total number of configuration validation errors",
		},
		[]string{"type", "error_type"},
	)
)
