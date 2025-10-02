package handlers

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/b25/services/payment/internal/logger"
	"github.com/yourorg/b25/services/payment/internal/service"
)

type WebhookHandler struct {
	service *service.WebhookService
	logger  *logger.Logger
}

func NewWebhookHandler(service *service.WebhookService, logger *logger.Logger) *WebhookHandler {
	return &WebhookHandler{
		service: service,
		logger:  logger,
	}
}

// HandleStripeWebhook handles Stripe webhook events
func (h *WebhookHandler) HandleStripeWebhook(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("Failed to read webhook payload", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	signature := c.GetHeader("Stripe-Signature")
	if signature == "" {
		h.logger.Error("Missing Stripe signature")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing signature"})
		return
	}

	if err := h.service.VerifyAndProcessWebhook(payload, signature); err != nil {
		h.logger.Error("Failed to process webhook", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Webhook processing failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}
