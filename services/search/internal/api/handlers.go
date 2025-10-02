package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/yourorg/b25/services/search/internal/analytics"
	"github.com/yourorg/b25/services/search/internal/config"
	"github.com/yourorg/b25/services/search/internal/search"
	"github.com/yourorg/b25/services/search/pkg/models"
)

// Handler handles HTTP requests
type Handler struct {
	es        *search.ElasticsearchClient
	analytics *analytics.Analytics
	config    *config.SearchConfig
	logger    *zap.Logger
}

// NewHandler creates a new HTTP handler
func NewHandler(
	es *search.ElasticsearchClient,
	analytics *analytics.Analytics,
	cfg *config.SearchConfig,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		es:        es,
		analytics: analytics,
		config:    cfg,
		logger:    logger,
	}
}

// Search handles search requests
// @Summary Search for documents
// @Description Performs full-text search across indices
// @Tags search
// @Accept json
// @Produce json
// @Param request body models.SearchRequest true "Search request"
// @Success 200 {object} models.SearchResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/search [post]
func (h *Handler) Search(c *gin.Context) {
	var req models.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	// Apply defaults
	if req.Size == 0 {
		req.Size = h.config.DefaultPageSize
	}
	if req.Size > h.config.MaxPageSize {
		req.Size = h.config.MaxPageSize
	}

	// Execute search
	ctx, cancel := context.WithTimeout(context.Background(), h.config.Timeout)
	defer cancel()

	startTime := time.Now()
	resp, err := h.es.Search(ctx, &req)
	if err != nil {
		h.logger.Error("Search failed",
			zap.Error(err),
			zap.String("query", req.Query),
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Search failed",
			Message: err.Error(),
		})
		return
	}

	// Track analytics
	go func() {
		searchAnalytics := &models.SearchAnalytics{
			Query:       req.Query,
			Index:       req.Index,
			ResultCount: resp.TotalHits,
			Timestamp:   time.Now(),
			Latency:     time.Since(startTime).Milliseconds(),
		}
		if err := h.analytics.TrackSearch(context.Background(), searchAnalytics); err != nil {
			h.logger.Warn("Failed to track search analytics", zap.Error(err))
		}
	}()

	c.JSON(http.StatusOK, resp)
}

// Autocomplete handles autocomplete requests
// @Summary Get autocomplete suggestions
// @Description Returns autocomplete suggestions for a query
// @Tags search
// @Accept json
// @Produce json
// @Param request body models.AutocompleteRequest true "Autocomplete request"
// @Success 200 {object} models.AutocompleteResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/autocomplete [post]
func (h *Handler) Autocomplete(c *gin.Context) {
	var req models.AutocompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	// Apply defaults
	if req.Size == 0 {
		req.Size = h.config.Autocomplete.MaxSuggestions
	}

	// Execute autocomplete
	ctx, cancel := context.WithTimeout(context.Background(), h.config.Timeout)
	defer cancel()

	resp, err := h.es.Autocomplete(ctx, &req)
	if err != nil {
		h.logger.Error("Autocomplete failed",
			zap.Error(err),
			zap.String("query", req.Query),
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Autocomplete failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Index handles document indexing requests
// @Summary Index a document
// @Description Indexes a single document
// @Tags indexing
// @Accept json
// @Produce json
// @Param request body models.IndexRequest true "Index request"
// @Success 200 {object} models.IndexResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/index [post]
func (h *Handler) Index(c *gin.Context) {
	var req models.IndexRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := h.es.Index(ctx, &req)
	if err != nil {
		h.logger.Error("Index failed",
			zap.Error(err),
			zap.String("index", req.Index),
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Index failed",
			Message: err.Error(),
		})
		return
	}

	if !resp.Success {
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// BulkIndex handles bulk indexing requests
// @Summary Bulk index documents
// @Description Indexes multiple documents in a single request
// @Tags indexing
// @Accept json
// @Produce json
// @Param request body models.BulkIndexRequest true "Bulk index request"
// @Success 200 {object} models.BulkIndexResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/index/bulk [post]
func (h *Handler) BulkIndex(c *gin.Context) {
	var req models.BulkIndexRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := h.es.BulkIndex(ctx, &req)
	if err != nil {
		h.logger.Error("Bulk index failed",
			zap.Error(err),
			zap.Int("count", len(req.Documents)),
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Bulk index failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// TrackClick handles click tracking requests
// @Summary Track a click on a search result
// @Description Records a click for analytics
// @Tags analytics
// @Accept json
// @Produce json
// @Param request body models.ClickAnalytics true "Click analytics"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/analytics/click [post]
func (h *Handler) TrackClick(c *gin.Context) {
	var req models.ClickAnalytics
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	req.Timestamp = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.analytics.TrackClick(ctx, &req); err != nil {
		h.logger.Error("Failed to track click", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to track click",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Click tracked",
	})
}

// GetPopularSearches returns popular search queries
// @Summary Get popular searches
// @Description Returns the most popular search queries
// @Tags analytics
// @Produce json
// @Param limit query int false "Number of results" default(10)
// @Success 200 {array} models.PopularSearch
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/analytics/popular [get]
func (h *Handler) GetPopularSearches(c *gin.Context) {
	limit := 10
	if l, ok := c.GetQuery("limit"); ok {
		if _, err := fmt.Sscanf(l, "%d", &limit); err != nil {
			limit = 10
		}
	}

	if limit > 100 {
		limit = 100
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	searches, err := h.analytics.GetPopularSearches(ctx, limit)
	if err != nil {
		h.logger.Error("Failed to get popular searches", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get popular searches",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, searches)
}

// GetSearchStats returns search statistics
// @Summary Get search statistics
// @Description Returns aggregated search statistics
// @Tags analytics
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/analytics/stats [get]
func (h *Handler) GetSearchStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stats, err := h.analytics.GetSearchStats(ctx)
	if err != nil {
		h.logger.Error("Failed to get search stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get search stats",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// Health handles health check requests
// @Summary Health check
// @Description Returns the health status of the service
// @Tags health
// @Produce json
// @Success 200 {object} models.HealthStatus
// @Failure 503 {object} models.HealthStatus
// @Router /health [get]
func (h *Handler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	health := &models.HealthStatus{
		Status:    "healthy",
		Version:   "1.0.0",
		Uptime:    getUptime(),
		Timestamp: time.Now(),
	}

	// Check Elasticsearch
	esHealth, err := h.es.Health(ctx)
	if err != nil {
		esHealth = &models.ComponentHealth{
			Status: "unhealthy",
			Error:  err.Error(),
		}
	}
	health.Elasticsearch = *esHealth

	// Determine overall health
	if esHealth.Status == "unhealthy" {
		health.Status = "unhealthy"
		c.JSON(http.StatusServiceUnavailable, health)
		return
	} else if esHealth.Status == "degraded" {
		health.Status = "degraded"
	}

	c.JSON(http.StatusOK, health)
}

// Readiness handles readiness probe requests
// @Summary Readiness check
// @Description Returns whether the service is ready to accept requests
// @Tags health
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 503 {object} ErrorResponse
// @Router /ready [get]
func (h *Handler) Readiness(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Quick Elasticsearch check
	esHealth, err := h.es.Health(ctx)
	if err != nil || esHealth.Status == "unhealthy" {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error:   "Service not ready",
			Message: "Elasticsearch is not available",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Service is ready",
	})
}

// Response types

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// Helper functions

var startTime = time.Now()

func getUptime() string {
	uptime := time.Since(startTime)
	return uptime.String()
}
