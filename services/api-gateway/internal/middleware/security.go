package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityMiddleware adds security headers to responses
type SecurityMiddleware struct {
	enableHSTS bool
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(enableHSTS bool) *SecurityMiddleware {
	return &SecurityMiddleware{
		enableHSTS: enableHSTS,
	}
}

// SecurityHeaders adds security headers to all responses
func (s *SecurityMiddleware) SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Enable XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")

		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Referrer policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy (basic policy, adjust as needed)
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self' wss: ws:;")

		// Permissions Policy (formerly Feature Policy)
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// HSTS (HTTP Strict Transport Security) - only if enabled and using HTTPS
		if s.enableHSTS && c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		// Remove server identification headers
		c.Header("Server", "")
		c.Header("X-Powered-By", "")

		c.Next()
	}
}

// RemoveServerHeaders removes server identification headers
func (s *SecurityMiddleware) RemoveServerHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Remove these headers after response
		c.Writer.Header().Del("Server")
		c.Writer.Header().Del("X-Powered-By")
	}
}
