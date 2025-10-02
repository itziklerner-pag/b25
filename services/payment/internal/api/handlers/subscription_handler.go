package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/b25/services/payment/internal/logger"
	"github.com/yourorg/b25/services/payment/internal/models"
	"github.com/yourorg/b25/services/payment/internal/service"
)

type SubscriptionHandler struct {
	service *service.SubscriptionService
	logger  *logger.Logger
}

func NewSubscriptionHandler(service *service.SubscriptionService, logger *logger.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		service: service,
		logger:  logger,
	}
}

// CreateSubscription creates a new subscription
func (h *SubscriptionHandler) CreateSubscription(c *gin.Context) {
	var req models.CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if exists {
		req.UserID = userID.(string)
	}

	subscription, err := h.service.CreateSubscription(&req)
	if err != nil {
		h.logger.Error("Failed to create subscription", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscription"})
		return
	}

	c.JSON(http.StatusCreated, subscription)
}

// GetSubscription retrieves a subscription by ID
func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	id := c.Param("id")

	subscription, err := h.service.GetSubscription(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// GetUserSubscriptions retrieves all subscriptions for a user
func (h *SubscriptionHandler) GetUserSubscriptions(c *gin.Context) {
	userID := c.Param("user_id")

	// Verify user has permission
	authUserID, _ := c.Get("user_id")
	if authUserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	subscriptions, err := h.service.GetUserSubscriptions(userID)
	if err != nil {
		h.logger.Error("Failed to get user subscriptions", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve subscriptions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": subscriptions})
}

// CancelSubscription cancels a subscription
func (h *SubscriptionHandler) CancelSubscription(c *gin.Context) {
	id := c.Param("id")

	// Check if immediate cancellation is requested
	immediate := c.Query("immediate") == "true"

	subscription, err := h.service.CancelSubscription(id, immediate)
	if err != nil {
		h.logger.Error("Failed to cancel subscription", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel subscription"})
		return
	}

	c.JSON(http.StatusOK, subscription)
}
