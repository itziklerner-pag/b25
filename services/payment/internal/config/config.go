package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Stripe   StripeConfig
	Security SecurityConfig
}

type ServerConfig struct {
	Port         int
	Environment  string
	ReadTimeout  int
	WriteTimeout int
	IdleTimeout  int
}

type DatabaseConfig struct {
	URL             string
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type StripeConfig struct {
	SecretKey     string
	PublishableKey string
	WebhookSecret string
	APIVersion    string
}

type SecurityConfig struct {
	JWTSecret           string
	EncryptionKey       string
	AllowedOrigins      []string
	RateLimitPerMinute  int
	EnableIPWhitelist   bool
	IPWhitelist         []string
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:         getEnvAsInt("SERVER_PORT", 9097),
			Environment:  getEnv("ENVIRONMENT", "development"),
			ReadTimeout:  getEnvAsInt("SERVER_READ_TIMEOUT", 30),
			WriteTimeout: getEnvAsInt("SERVER_WRITE_TIMEOUT", 30),
			IdleTimeout:  getEnvAsInt("SERVER_IDLE_TIMEOUT", 120),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", ""),
			Database:        getEnv("DB_NAME", "payment_db"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsInt("DB_CONN_MAX_LIFETIME", 300),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Stripe: StripeConfig{
			SecretKey:      getEnv("STRIPE_SECRET_KEY", ""),
			PublishableKey: getEnv("STRIPE_PUBLISHABLE_KEY", ""),
			WebhookSecret:  getEnv("STRIPE_WEBHOOK_SECRET", ""),
			APIVersion:     getEnv("STRIPE_API_VERSION", "2023-10-16"),
		},
		Security: SecurityConfig{
			JWTSecret:          getEnv("JWT_SECRET", ""),
			EncryptionKey:      getEnv("ENCRYPTION_KEY", ""),
			AllowedOrigins:     getEnvAsSlice("ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
			RateLimitPerMinute: getEnvAsInt("RATE_LIMIT_PER_MINUTE", 100),
			EnableIPWhitelist:  getEnvAsBool("ENABLE_IP_WHITELIST", false),
			IPWhitelist:        getEnvAsSlice("IP_WHITELIST", []string{}),
		},
	}

	// Build database URL
	cfg.Database.URL = fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Database,
		cfg.Database.SSLMode,
	)

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Stripe.SecretKey == "" {
		return fmt.Errorf("STRIPE_SECRET_KEY is required")
	}
	if c.Stripe.WebhookSecret == "" {
		return fmt.Errorf("STRIPE_WEBHOOK_SECRET is required")
	}
	if c.Security.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if c.Security.EncryptionKey == "" {
		return fmt.Errorf("ENCRYPTION_KEY is required")
	}
	if len(c.Security.EncryptionKey) != 32 {
		return fmt.Errorf("ENCRYPTION_KEY must be exactly 32 characters")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return splitString(value, ",")
	}
	return defaultValue
}

func splitString(s, sep string) []string {
	var result []string
	for _, item := range splitByRune(s, ',') {
		if trimmed := trimSpace(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func splitByRune(s string, r rune) []string {
	var result []string
	var current string
	for _, char := range s {
		if char == r {
			result = append(result, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n') {
		end--
	}
	return s[start:end]
}
