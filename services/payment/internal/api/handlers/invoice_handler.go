package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/b25/services/payment/internal/logger"
	"github.com/yourorg/b25/services/payment/internal/service"
)

type InvoiceHandler struct {
	service *service.InvoiceService
	logger  *logger.Logger
}

func NewInvoiceHandler(service *service.InvoiceService, logger *logger.Logger) *InvoiceHandler {
	return &InvoiceHandler{
		service: service,
		logger:  logger,
	}
}

// GetInvoice retrieves an invoice by ID
func (h *InvoiceHandler) GetInvoice(c *gin.Context) {
	id := c.Param("id")

	invoice, err := h.service.GetInvoice(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
		return
	}

	c.JSON(http.StatusOK, invoice)
}

// GetUserInvoices retrieves all invoices for a user
func (h *InvoiceHandler) GetUserInvoices(c *gin.Context) {
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

	invoices, err := h.service.GetUserInvoices(userID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get user invoices", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve invoices"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   invoices,
		"limit":  limit,
		"offset": offset,
	})
}

// GetSubscriptionInvoices retrieves all invoices for a subscription
func (h *InvoiceHandler) GetSubscriptionInvoices(c *gin.Context) {
	subscriptionID := c.Param("subscription_id")

	invoices, err := h.service.GetSubscriptionInvoices(subscriptionID)
	if err != nil {
		h.logger.Error("Failed to get subscription invoices", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve invoices"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": invoices})
}
