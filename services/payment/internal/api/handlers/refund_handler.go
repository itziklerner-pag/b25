package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/b25/services/payment/internal/logger"
	"github.com/yourorg/b25/services/payment/internal/models"
	"github.com/yourorg/b25/services/payment/internal/service"
)

type RefundHandler struct {
	service *service.RefundService
	logger  *logger.Logger
}

func NewRefundHandler(service *service.RefundService, logger *logger.Logger) *RefundHandler {
	return &RefundHandler{
		service: service,
		logger:  logger,
	}
}

// CreateRefund creates a new refund
func (h *RefundHandler) CreateRefund(c *gin.Context) {
	var req models.CreateRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	refund, err := h.service.CreateRefund(&req)
	if err != nil {
		h.logger.Error("Failed to create refund", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, refund)
}
