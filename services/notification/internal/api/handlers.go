package api

import (
	"net/http"

	"github.com/b25/services/notification/internal/models"
	"github.com/b25/services/notification/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// NotificationHandler handles notification-related HTTP requests
type NotificationHandler struct {
	service *service.NotificationService
	logger  *zap.Logger
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(service *service.NotificationService, logger *zap.Logger) *NotificationHandler {
	return &NotificationHandler{
		service: service,
		logger:  logger,
	}
}

// CreateNotification creates a new notification
// @Summary Create notification
// @Tags notifications
// @Accept json
// @Produce json
// @Param notification body models.CreateNotificationRequest true "Notification request"
// @Success 201 {object} models.Notification
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/notifications [post]
func (h *NotificationHandler) CreateNotification(c *gin.Context) {
	var req models.CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	notification, err := h.service.CreateNotification(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("failed to create notification", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, notification)
}

// GetNotification retrieves a notification by ID
// @Summary Get notification
// @Tags notifications
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} models.Notification
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/notifications/{id} [get]
func (h *NotificationHandler) GetNotification(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid notification ID"})
		return
	}

	notification, err := h.service.GetNotification(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("failed to get notification", zap.Error(err))
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "notification not found"})
		return
	}

	c.JSON(http.StatusOK, notification)
}

// ListNotifications lists notifications with filters
// @Summary List notifications
// @Tags notifications
// @Produce json
// @Param user_id query string false "User ID"
// @Param channel query string false "Channel"
// @Param status query string false "Status"
// @Param priority query string false "Priority"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} models.PaginatedResult
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/notifications [get]
func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	filter := &models.QueryFilter{
		Limit:  c.GetInt("limit"),
		Offset: c.GetInt("offset"),
	}

	if filter.Limit == 0 {
		filter.Limit = 20
	}

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err == nil {
			filter.UserID = &userID
		}
	}

	if channelStr := c.Query("channel"); channelStr != "" {
		channel := models.NotificationChannel(channelStr)
		filter.Channel = &channel
	}

	if statusStr := c.Query("status"); statusStr != "" {
		status := models.NotificationStatus(statusStr)
		filter.Status = &status
	}

	if priorityStr := c.Query("priority"); priorityStr != "" {
		priority := models.NotificationPriority(priorityStr)
		filter.Priority = &priority
	}

	result, err := h.service.ListNotifications(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("failed to list notifications", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetUserNotifications retrieves notifications for a specific user
// @Summary Get user notifications
// @Tags notifications
// @Produce json
// @Param user_id path string true "User ID"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} models.PaginatedResult
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/notifications/user/{user_id} [get]
func (h *NotificationHandler) GetUserNotifications(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user ID"})
		return
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		limit = c.GetInt("limit")
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		offset = c.GetInt("offset")
	}

	result, err := h.service.GetUserNotifications(c.Request.Context(), userID, limit, offset)
	if err != nil {
		h.logger.Error("failed to get user notifications", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetByCorrelationID retrieves notifications by correlation ID
// @Summary Get notifications by correlation ID
// @Tags notifications
// @Produce json
// @Param correlation_id path string true "Correlation ID"
// @Success 200 {array} models.Notification
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/notifications/correlation/{correlation_id} [get]
func (h *NotificationHandler) GetByCorrelationID(c *gin.Context) {
	correlationID := c.Param("correlation_id")

	// This would need to be implemented in the service
	// For now, return empty array
	c.JSON(http.StatusOK, []models.Notification{})
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// HealthCheck handles health check requests
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "notification-service",
	})
}

// ReadinessCheck checks if the service is ready to serve requests
func ReadinessCheck(service *service.NotificationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Add actual readiness checks (database, redis, etc.)
		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
		})
	}
}
