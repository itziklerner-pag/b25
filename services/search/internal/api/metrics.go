package api

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "search_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "search_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Search metrics
	searchQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "search_queries_total",
			Help: "Total number of search queries",
		},
		[]string{"index"},
	)

	searchQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "search_query_duration_seconds",
			Help:    "Search query duration in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
		},
		[]string{"index"},
	)

	searchResultsCount = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "search_results_count",
			Help:    "Number of search results returned",
			Buckets: []float64{0, 1, 5, 10, 25, 50, 100, 500, 1000, 5000, 10000},
		},
		[]string{"index"},
	)

	// Indexing metrics
	indexOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "search_index_operations_total",
			Help: "Total number of index operations",
		},
		[]string{"index", "operation"},
	)

	indexBatchSize = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "search_index_batch_size",
			Help:    "Size of index batches",
			Buckets: []float64{1, 10, 50, 100, 500, 1000, 5000},
		},
	)

	indexDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "search_index_duration_seconds",
			Help:    "Index operation duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// Queue metrics
	indexQueueSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "search_index_queue_size",
			Help: "Current size of the index queue",
		},
	)

	indexQueueCapacity = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "search_index_queue_capacity",
			Help: "Capacity of the index queue",
		},
	)

	// Analytics metrics
	analyticsEventsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "search_analytics_events_total",
			Help: "Total number of analytics events",
		},
		[]string{"event_type"},
	)

	// Elasticsearch metrics
	elasticsearchHealthStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "search_elasticsearch_health_status",
			Help: "Elasticsearch health status (0=unhealthy, 1=degraded, 2=healthy)",
		},
		[]string{"cluster"},
	)

	elasticsearchRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "search_elasticsearch_requests_total",
			Help: "Total number of Elasticsearch requests",
		},
		[]string{"operation", "status"},
	)

	// Cache metrics
	cacheHitsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "search_cache_hits_total",
			Help: "Total number of cache hits",
		},
	)

	cacheMissesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "search_cache_misses_total",
			Help: "Total number of cache misses",
		},
	)
)

// RecordSearchQuery records a search query metric
func RecordSearchQuery(index string, duration float64, resultCount int64) {
	searchQueriesTotal.WithLabelValues(index).Inc()
	searchQueryDuration.WithLabelValues(index).Observe(duration)
	searchResultsCount.WithLabelValues(index).Observe(float64(resultCount))
}

// RecordIndexOperation records an indexing operation metric
func RecordIndexOperation(index, operation string, duration float64, batchSize int) {
	indexOperationsTotal.WithLabelValues(index, operation).Inc()
	indexDuration.WithLabelValues(operation).Observe(duration)
	if batchSize > 0 {
		indexBatchSize.Observe(float64(batchSize))
	}
}

// UpdateQueueSize updates the queue size metric
func UpdateQueueSize(size, capacity int) {
	indexQueueSize.Set(float64(size))
	indexQueueCapacity.Set(float64(capacity))
}

// RecordAnalyticsEvent records an analytics event
func RecordAnalyticsEvent(eventType string) {
	analyticsEventsTotal.WithLabelValues(eventType).Inc()
}

// UpdateElasticsearchHealth updates Elasticsearch health metric
func UpdateElasticsearchHealth(cluster, status string) {
	var value float64
	switch status {
	case "healthy":
		value = 2
	case "degraded":
		value = 1
	default:
		value = 0
	}
	elasticsearchHealthStatus.WithLabelValues(cluster).Set(value)
}

// RecordElasticsearchRequest records an Elasticsearch request
func RecordElasticsearchRequest(operation, status string) {
	elasticsearchRequestsTotal.WithLabelValues(operation, status).Inc()
}

// RecordCacheHit records a cache hit
func RecordCacheHit() {
	cacheHitsTotal.Inc()
}

// RecordCacheMiss records a cache miss
func RecordCacheMiss() {
	cacheMissesTotal.Inc()
}
