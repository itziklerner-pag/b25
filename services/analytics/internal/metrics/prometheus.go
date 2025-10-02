package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics
type Metrics struct {
	EventsIngested   prometheus.Counter
	EventsFailed     prometheus.Counter
	BatchesProcessed prometheus.Counter
	BatchDuration    prometheus.Histogram
	QueryDuration    prometheus.Histogram
	CacheHits        prometheus.Counter
	CacheMisses      prometheus.Counter
	ActiveConnections prometheus.Gauge
}

// NewMetrics creates a new Metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		EventsIngested: promauto.NewCounter(prometheus.CounterOpts{
			Name: "analytics_events_ingested_total",
			Help: "Total number of events ingested",
		}),
		EventsFailed: promauto.NewCounter(prometheus.CounterOpts{
			Name: "analytics_events_failed_total",
			Help: "Total number of events that failed to ingest",
		}),
		BatchesProcessed: promauto.NewCounter(prometheus.CounterOpts{
			Name: "analytics_batches_processed_total",
			Help: "Total number of batches processed",
		}),
		BatchDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "analytics_batch_duration_seconds",
			Help:    "Duration of batch processing in seconds",
			Buckets: prometheus.DefBuckets,
		}),
		QueryDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "analytics_query_duration_seconds",
			Help:    "Duration of database queries in seconds",
			Buckets: prometheus.DefBuckets,
		}),
		CacheHits: promauto.NewCounter(prometheus.CounterOpts{
			Name: "analytics_cache_hits_total",
			Help: "Total number of cache hits",
		}),
		CacheMisses: promauto.NewCounter(prometheus.CounterOpts{
			Name: "analytics_cache_misses_total",
			Help: "Total number of cache misses",
		}),
		ActiveConnections: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "analytics_active_connections",
			Help: "Number of active connections",
		}),
	}
}
