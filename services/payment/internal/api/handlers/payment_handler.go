package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/b25/services/payment/internal/logger"
	"github.com/yourorg/b25/services/payment/internal/models"
	"github.com/yourorg/b25/services/payment/internal/service"
)

type PaymentHandler struct {
	service *service.PaymentService
	logger  *logger.Logger
}

func NewPaymentHandler(service *service.PaymentService, logger *logger.Logger) *PaymentHandler {
	return &PaymentHandler{
		service: service,
		logger:  logger,
	}
}

// CreatePayment creates a new payment
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req models.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if exists {
		req.UserID = userID.(string)
	}

	tx, err := h.service.CreatePayment(&req)
	if err != nil {
		h.logger.Error("Failed to create payment", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment"})
		return
	}

	c.JSON(http.StatusCreated, tx)
}

// GetPayment retrieves a payment by ID
func (h *PaymentHandler) GetPayment(c *gin.Context) {
	id := c.Param("id")

	tx, err := h.service.GetTransaction(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}

	c.JSON(http.StatusOK, tx)
}

// GetUserPayments retrieves all payments for a user
func (h *PaymentHandler) GetUserPayments(c *gin.Context) {
	userID := c.Param("user_id")

	// Verify user has permission
	authUserID, _ := c.Get("user_id")
	if authUserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Parse pagination parameters
	limit := 20
	offset := 0
	if l := c.Query("limit"); l != "" {
		if _, err := parseIntQuery(l, &limit); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
			return
		}
	}
	if o := c.Query("offset"); o != "" {
		if _, err := parseIntQuery(o, &offset); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset parameter"})
			return
		}
	}

	transactions, err := h.service.GetUserTransactions(userID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get user payments", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve payments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   transactions,
		"limit":  limit,
		"offset": offset,
	})
}

// AttachPaymentMethod attaches a payment method to a user
func (h *PaymentHandler) AttachPaymentMethod(c *gin.Context) {
	var req models.PaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if exists {
		req.UserID = userID.(string)
	}

	pm, err := h.service.AttachPaymentMethod(&req)
	if err != nil {
		h.logger.Error("Failed to attach payment method", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to attach payment method"})
		return
	}

	c.JSON(http.StatusCreated, pm)
}

// GetUserPaymentMethods retrieves all payment methods for a user
func (h *PaymentHandler) GetUserPaymentMethods(c *gin.Context) {
	userID := c.Param("user_id")

	// Verify user has permission
	authUserID, _ := c.Get("user_id")
	if authUserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	paymentMethods, err := h.service.GetUserPaymentMethods(userID)
	if err != nil {
		h.logger.Error("Failed to get user payment methods", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve payment methods"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": paymentMethods})
}

// DetachPaymentMethod removes a payment method
func (h *PaymentHandler) DetachPaymentMethod(c *gin.Context) {
	id := c.Param("id")

	// Get user ID from context
	userID, _ := c.Get("user_id")

	if err := h.service.DetachPaymentMethod(id, userID.(string)); err != nil {
		h.logger.Error("Failed to detach payment method", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to detach payment method"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment method removed successfully"})
}
