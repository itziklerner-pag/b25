package api

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// APIKeyMiddleware validates API key in Authorization header or X-API-Key header
func APIKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get API key from environment variable
		validAPIKey := os.Getenv("CONFIG_API_KEY")

		// If no API key is configured, allow all requests (backward compatibility)
		if validAPIKey == "" {
			c.Next()
			return
		}

		// Try to get API key from Authorization header (Bearer token format)
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			// Expected format: "Bearer <api-key>"
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				if parts[1] == validAPIKey {
					c.Next()
					return
				}
			}
		}

		// Try to get API key from X-API-Key header
		apiKey := c.GetHeader("X-API-Key")
		if apiKey != "" && apiKey == validAPIKey {
			c.Next()
			return
		}

		// No valid API key found
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Error:   "Unauthorized: Invalid or missing API key",
			Code:    "UNAUTHORIZED",
		})
		c.Abort()
	}
}
