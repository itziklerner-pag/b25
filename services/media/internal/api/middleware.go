package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware configures CORS
func CORSMiddleware() gin.HandlerFunc {
	config := cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-User-ID", "X-Org-ID"},
		ExposeHeaders:    []string{"Content-Length", "Content-Range", "Content-Type"},
		AllowCredentials: true,
	}

	return cors.New(config)
}

// AuthMiddleware validates authentication (placeholder)
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// In a real application, this would validate JWT tokens or API keys
		// For now, we'll extract user_id and org_id from headers

		userID := c.GetHeader("X-User-ID")
		orgID := c.GetHeader("X-Org-ID")

		if userID != "" {
			c.Set("user_id", userID)
		}

		if orgID != "" {
			c.Set("org_id", orgID)
		}

		c.Next()
	}
}

// RateLimitMiddleware implements rate limiting (placeholder)
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// In a real application, this would implement rate limiting
		// using Redis or an in-memory cache
		c.Next()
	}
}
