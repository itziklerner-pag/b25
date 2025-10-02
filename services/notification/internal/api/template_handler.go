package api

import (
	"net/http"

	"github.com/b25/services/notification/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// TemplateHandler handles template-related HTTP requests
type TemplateHandler struct {
	service *service.TemplateService
	logger  *zap.Logger
}

// NewTemplateHandler creates a new template handler
func NewTemplateHandler(service *service.TemplateService, logger *zap.Logger) *TemplateHandler {
	return &TemplateHandler{
		service: service,
		logger:  logger,
	}
}

// CreateTemplate creates a new template
func (h *TemplateHandler) CreateTemplate(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, ErrorResponse{Error: "not implemented"})
}

// GetTemplate retrieves a template by ID
func (h *TemplateHandler) GetTemplate(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, ErrorResponse{Error: "not implemented"})
}

// ListTemplates lists all templates
func (h *TemplateHandler) ListTemplates(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, ErrorResponse{Error: "not implemented"})
}

// UpdateTemplate updates a template
func (h *TemplateHandler) UpdateTemplate(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, ErrorResponse{Error: "not implemented"})
}

// DeleteTemplate deletes a template
func (h *TemplateHandler) DeleteTemplate(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, ErrorResponse{Error: "not implemented"})
}

// TestTemplate tests a template with sample data
func (h *TemplateHandler) TestTemplate(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, ErrorResponse{Error: "not implemented"})
}
