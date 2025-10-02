package api

import (
	"github.com/b25/services/content/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	contentService service.ContentService
	authService    service.AuthService
	mediaService   service.MediaService
	logger         *zap.Logger
}

func NewHandler(
	contentService service.ContentService,
	authService service.AuthService,
	mediaService service.MediaService,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		contentService: contentService,
		authService:    authService,
		mediaService:   mediaService,
		logger:         logger,
	}
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse represents a successful response
type SuccessResponse struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}

// sendError sends an error response
func (h *Handler) sendError(c *gin.Context, statusCode int, err error, message string) {
	response := ErrorResponse{
		Error:   err.Error(),
		Message: message,
	}
	c.JSON(statusCode, response)
}

// sendSuccess sends a success response
func (h *Handler) sendSuccess(c *gin.Context, statusCode int, data interface{}, message string) {
	response := SuccessResponse{
		Data:    data,
		Message: message,
	}
	c.JSON(statusCode, response)
}
