package middleware

import (
	"net/http"

	"github.com/b25/api-gateway/internal/config"
	"github.com/b25/api-gateway/pkg/logger"
	"github.com/gin-gonic/gin"
)

// ValidationMiddleware handles request validation
type ValidationMiddleware struct {
	config config.ValidationConfig
	log    *logger.Logger
}

// NewValidationMiddleware creates a new validation middleware
func NewValidationMiddleware(cfg config.ValidationConfig, log *logger.Logger) *ValidationMiddleware {
	return &ValidationMiddleware{
		config: cfg,
		log:    log,
	}
}

// ValidateRequestSize validates request body size
func (v *ValidationMiddleware) ValidateRequestSize() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !v.config.Enabled {
			c.Next()
			return
		}

		if c.Request.ContentLength > v.config.MaxRequestSize {
			v.log.Warn("Request body too large",
				"size", c.Request.ContentLength,
				"max_size", v.config.MaxRequestSize,
				"path", c.Request.URL.Path,
			)
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "Request body too large",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateContentType validates request content type
func (v *ValidationMiddleware) ValidateContentType(allowedTypes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" || c.Request.Method == "DELETE" {
			c.Next()
			return
		}

		contentType := c.ContentType()
		if contentType == "" {
			c.Next()
			return
		}

		allowed := false
		for _, allowedType := range allowedTypes {
			if contentType == allowedType {
				allowed = true
				break
			}
		}

		if !allowed {
			v.log.Warn("Invalid content type",
				"content_type", contentType,
				"allowed_types", allowedTypes,
				"path", c.Request.URL.Path,
			)
			c.JSON(http.StatusUnsupportedMediaType, gin.H{
				"error": "Unsupported content type",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateHeaders validates required headers
func (v *ValidationMiddleware) ValidateHeaders(requiredHeaders ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, header := range requiredHeaders {
			if c.GetHeader(header) == "" {
				v.log.Warn("Missing required header",
					"header", header,
					"path", c.Request.URL.Path,
				)
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Missing required header: " + header,
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
