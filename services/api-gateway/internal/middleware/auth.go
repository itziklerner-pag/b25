package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/b25/api-gateway/internal/config"
	"github.com/b25/api-gateway/pkg/logger"
	"github.com/b25/api-gateway/pkg/metrics"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	AuthTypeJWT    = "jwt"
	AuthTypeAPIKey = "api_key"
)

// AuthMiddleware handles authentication
type AuthMiddleware struct {
	config  config.AuthConfig
	log     *logger.Logger
	metrics *metrics.Collector
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(cfg config.AuthConfig, log *logger.Logger, m *metrics.Collector) *AuthMiddleware {
	return &AuthMiddleware{
		config:  cfg,
		log:     log,
		metrics: m,
	}
}

// Authenticate is a middleware that validates authentication
func (a *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !a.config.Enabled {
			c.Next()
			return
		}

		// Try JWT authentication first
		token := extractBearerToken(c)
		if token != "" {
			if err := a.validateJWT(token); err != nil {
				a.metrics.RecordAuthAttempt(AuthTypeJWT, false, "invalid_token")
				a.log.Warn("JWT authentication failed", "error", err, "path", c.Request.URL.Path)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
				c.Abort()
				return
			}
			a.metrics.RecordAuthAttempt(AuthTypeJWT, true, "")
			c.Set("auth_type", AuthTypeJWT)
			c.Next()
			return
		}

		// Try API key authentication
		apiKey := extractAPIKey(c)
		if apiKey != "" {
			role, err := a.validateAPIKey(apiKey)
			if err != nil {
				a.metrics.RecordAuthAttempt(AuthTypeAPIKey, false, "invalid_key")
				a.log.Warn("API key authentication failed", "error", err, "path", c.Request.URL.Path)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
				c.Abort()
				return
			}
			a.metrics.RecordAuthAttempt(AuthTypeAPIKey, true, "")
			c.Set("auth_type", AuthTypeAPIKey)
			c.Set("user_role", role)
			c.Next()
			return
		}

		// No authentication provided
		a.metrics.RecordAuthAttempt("none", false, "missing_credentials")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		c.Abort()
	}
}

// OptionalAuth is middleware for optional authentication
func (a *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !a.config.Enabled {
			c.Next()
			return
		}

		// Try JWT
		token := extractBearerToken(c)
		if token != "" {
			if err := a.validateJWT(token); err == nil {
				c.Set("auth_type", AuthTypeJWT)
				c.Set("authenticated", true)
			}
		}

		// Try API key
		apiKey := extractAPIKey(c)
		if apiKey != "" {
			if role, err := a.validateAPIKey(apiKey); err == nil {
				c.Set("auth_type", AuthTypeAPIKey)
				c.Set("user_role", role)
				c.Set("authenticated", true)
			}
		}

		c.Next()
	}
}

// RequireRole checks if the user has the required role
func (a *AuthMiddleware) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid role"})
			c.Abort()
			return
		}

		for _, requiredRole := range roles {
			if role == requiredRole || role == "admin" { // admin has access to everything
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		c.Abort()
	}
}

// validateJWT validates a JWT token
func (a *AuthMiddleware) validateJWT(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.config.JWTSecret), nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("invalid claims")
	}

	// Check expiration
	if exp, ok := claims["exp"].(float64); ok {
		if time.Unix(int64(exp), 0).Before(time.Now()) {
			return fmt.Errorf("token expired")
		}
	}

	return nil
}

// validateAPIKey validates an API key
func (a *AuthMiddleware) validateAPIKey(key string) (string, error) {
	for _, apiKey := range a.config.APIKeys {
		if apiKey.Key == key {
			return apiKey.Role, nil
		}
	}
	return "", fmt.Errorf("invalid API key")
}

// extractBearerToken extracts Bearer token from Authorization header
func extractBearerToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		return ""
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

// extractAPIKey extracts API key from X-API-Key header
func extractAPIKey(c *gin.Context) string {
	return c.GetHeader("X-API-Key")
}

// GenerateJWT generates a new JWT token
func (a *AuthMiddleware) GenerateJWT(userID, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(a.config.JWTExpiry).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.config.JWTSecret))
}
