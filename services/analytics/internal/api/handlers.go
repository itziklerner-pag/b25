package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/b25/analytics/internal/cache"
	"github.com/b25/analytics/internal/models"
	"github.com/b25/analytics/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Handler handles HTTP requests
type Handler struct {
	repo   *repository.Repository
	cache  *cache.RedisCache
	logger *zap.Logger
}

// NewHandler creates a new API handler
func NewHandler(repo *repository.Repository, cache *cache.RedisCache, logger *zap.Logger) *Handler {
	return &Handler{
		repo:   repo,
		cache:  cache,
		logger: logger,
	}
}

// TrackEvent handles event tracking requests
func (h *Handler) TrackEvent(c *gin.Context) {
	var event models.Event
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Set defaults
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	event.CreatedAt = time.Now()

	if event.Properties == nil {
		event.Properties = make(map[string]interface{})
	}

	// Insert event
	if err := h.repo.InsertEvent(c.Request.Context(), &event); err != nil {
		h.logger.Error("Failed to insert event", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track event"})
		return
	}

	// Update cache counters
	if err := h.cache.IncrementEventCounter(c.Request.Context(), event.EventType); err != nil {
		h.logger.Warn("Failed to update event counter", zap.Error(err))
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"event_id": event.ID,
	})
}

// GetEvents retrieves events within a time range
func (h *Handler) GetEvents(c *gin.Context) {
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")
	eventType := c.Query("event_type")
	limitStr := c.DefaultQuery("limit", "100")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 10000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	var startTime, endTime time.Time
	if startTimeStr != "" {
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_time format"})
			return
		}
	} else {
		startTime = time.Now().Add(-24 * time.Hour)
	}

	if endTimeStr != "" {
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_time format"})
			return
		}
	} else {
		endTime = time.Now()
	}

	events, err := h.repo.GetEventsByTimeRange(c.Request.Context(), startTime, endTime, eventType, limit)
	if err != nil {
		h.logger.Error("Failed to get events", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"count":  len(events),
		"start_time": startTime,
		"end_time": endTime,
	})
}

// GetMetrics retrieves aggregated metrics
func (h *Handler) GetMetrics(c *gin.Context) {
	metricName := c.Query("metric_name")
	interval := c.DefaultQuery("interval", "1h")
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	if metricName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "metric_name is required"})
		return
	}

	var startTime, endTime time.Time
	var err error

	if startTimeStr != "" {
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_time format"})
			return
		}
	} else {
		startTime = time.Now().Add(-24 * time.Hour)
	}

	if endTimeStr != "" {
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_time format"})
			return
		}
	} else {
		endTime = time.Now()
	}

	// Check cache first
	cacheKey := cache.GenerateCacheKey("metrics", map[string]interface{}{
		"metric_name": metricName,
		"interval":    interval,
		"start_time":  startTime.Unix(),
		"end_time":    endTime.Unix(),
	})

	cachedResult, err := h.cache.GetQueryResult(c.Request.Context(), cacheKey)
	if err != nil {
		h.logger.Warn("Failed to get from cache", zap.Error(err))
	} else if cachedResult != nil {
		c.JSON(http.StatusOK, gin.H{
			"cached": true,
			"result": cachedResult,
		})
		return
	}

	// Query database
	aggregations, err := h.repo.GetMetricAggregations(c.Request.Context(), metricName, interval, startTime, endTime)
	if err != nil {
		h.logger.Error("Failed to get metrics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve metrics"})
		return
	}

	// Convert to query result
	dataPoints := make([]models.TimeSeriesPoint, len(aggregations))
	for i, agg := range aggregations {
		value := float64(agg.Count)
		if agg.Avg != nil {
			value = *agg.Avg
		}

		dataPoints[i] = models.TimeSeriesPoint{
			Timestamp:  agg.TimeBucket,
			Value:      value,
			Dimensions: agg.Dimensions,
		}
	}

	result := &models.QueryResult{
		MetricName:  metricName,
		Interval:    interval,
		StartTime:   startTime,
		EndTime:     endTime,
		DataPoints:  dataPoints,
		TotalCount:  len(dataPoints),
		Aggregation: "avg",
	}

	// Cache the result
	if err := h.cache.SetQueryResult(c.Request.Context(), cacheKey, result); err != nil {
		h.logger.Warn("Failed to cache result", zap.Error(err))
	}

	c.JSON(http.StatusOK, gin.H{
		"cached": false,
		"result": result,
	})
}

// GetDashboardMetrics retrieves real-time dashboard metrics
func (h *Handler) GetDashboardMetrics(c *gin.Context) {
	// Try cache first
	metrics, err := h.cache.GetDashboardMetrics(c.Request.Context())
	if err != nil {
		h.logger.Warn("Failed to get dashboard metrics from cache", zap.Error(err))
	}

	// If not in cache or cache miss, query database
	if metrics == nil {
		metrics, err = h.repo.GetDashboardMetrics(c.Request.Context())
		if err != nil {
			h.logger.Error("Failed to get dashboard metrics", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve dashboard metrics"})
			return
		}

		// Cache the result
		if err := h.cache.SetDashboardMetrics(c.Request.Context(), metrics); err != nil {
			h.logger.Warn("Failed to cache dashboard metrics", zap.Error(err))
		}
	}

	c.JSON(http.StatusOK, metrics)
}

// GetEventStats retrieves event statistics
func (h *Handler) GetEventStats(c *gin.Context) {
	startTimeStr := c.DefaultQuery("start_time", "")
	endTimeStr := c.DefaultQuery("end_time", "")

	var startTime, endTime time.Time
	var err error

	if startTimeStr != "" {
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_time format"})
			return
		}
	} else {
		startTime = time.Now().Add(-24 * time.Hour)
	}

	if endTimeStr != "" {
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_time format"})
			return
		}
	} else {
		endTime = time.Now()
	}

	counts, err := h.repo.GetEventCountByType(c.Request.Context(), startTime, endTime)
	if err != nil {
		h.logger.Error("Failed to get event stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve event statistics"})
		return
	}

	var totalEvents int64
	for _, count := range counts {
		totalEvents += count
	}

	c.JSON(http.StatusOK, gin.H{
		"start_time":   startTime,
		"end_time":     endTime,
		"total_events": totalEvents,
		"by_type":      counts,
	})
}

// CreateCustomEvent creates a custom event definition
func (h *Handler) CreateCustomEvent(c *gin.Context) {
	var def models.CustomEventDefinition
	if err := c.ShouldBindJSON(&def); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	def.ID = uuid.New()
	def.CreatedAt = time.Now()
	def.UpdatedAt = time.Now()

	if err := h.repo.CreateCustomEventDefinition(c.Request.Context(), &def); err != nil {
		h.logger.Error("Failed to create custom event definition", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create custom event"})
		return
	}

	c.JSON(http.StatusCreated, def)
}

// GetCustomEvent retrieves a custom event definition
func (h *Handler) GetCustomEvent(c *gin.Context) {
	name := c.Param("name")

	def, err := h.repo.GetCustomEventDefinition(c.Request.Context(), name)
	if err != nil {
		h.logger.Error("Failed to get custom event definition", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Custom event not found"})
		return
	}

	c.JSON(http.StatusOK, def)
}

// HealthCheck returns service health status
func (h *Handler) HealthCheck(c *gin.Context) {
	health := gin.H{
		"status": "healthy",
		"timestamp": time.Now(),
		"service": "analytics",
	}

	// Check database
	if err := h.repo.GetHealth(c.Request.Context()); err != nil {
		health["status"] = "unhealthy"
		health["database"] = "disconnected"
		c.JSON(http.StatusServiceUnavailable, health)
		return
	}
	health["database"] = "connected"

	// Check Redis
	if err := h.cache.GetHealth(c.Request.Context()); err != nil {
		health["status"] = "degraded"
		health["cache"] = "disconnected"
	} else {
		health["cache"] = "connected"
	}

	statusCode := http.StatusOK
	if health["status"] == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, health)
}

// GetIngestionMetrics returns ingestion metrics (would be wired to the consumer)
func (h *Handler) GetIngestionMetrics(c *gin.Context) {
	// This would be connected to the actual consumer metrics
	c.JSON(http.StatusOK, gin.H{
		"events_ingested": 0,
		"events_failed":   0,
		"batches_processed": 0,
	})
}
