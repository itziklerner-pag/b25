package middleware

import (
	"fmt"

	"github.com/b25/api-gateway/internal/config"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware handles CORS configuration
type CORSMiddleware struct {
	config config.CORSConfig
}

// NewCORSMiddleware creates a new CORS middleware
func NewCORSMiddleware(cfg config.CORSConfig) *CORSMiddleware {
	return &CORSMiddleware{
		config: cfg,
	}
}

// Handle applies CORS headers to responses
func (m *CORSMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.config.Enabled {
			c.Next()
			return
		}

		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range m.config.AllowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}

		if !allowed && len(m.config.AllowedOrigins) > 0 {
			origin = m.config.AllowedOrigins[0]
		}

		// Set CORS headers
		c.Header("Access-Control-Allow-Origin", origin)

		if m.config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if len(m.config.AllowedMethods) > 0 {
			methods := ""
			for i, method := range m.config.AllowedMethods {
				if i > 0 {
					methods += ", "
				}
				methods += method
			}
			c.Header("Access-Control-Allow-Methods", methods)
		}

		if len(m.config.AllowedHeaders) > 0 {
			headers := ""
			for i, header := range m.config.AllowedHeaders {
				if i > 0 {
					headers += ", "
				}
				headers += header
			}
			c.Header("Access-Control-Allow-Headers", headers)
		}

		if len(m.config.ExposedHeaders) > 0 {
			headers := ""
			for i, header := range m.config.ExposedHeaders {
				if i > 0 {
					headers += ", "
				}
				headers += header
			}
			c.Header("Access-Control-Expose-Headers", headers)
		}

		if m.config.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", m.config.MaxAge))
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
