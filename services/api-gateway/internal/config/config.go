package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the complete application configuration
type Config struct {
	Server         ServerConfig         `yaml:"server"`
	Services       ServicesConfig       `yaml:"services"`
	Auth           AuthConfig           `yaml:"auth"`
	RateLimit      RateLimitConfig      `yaml:"rate_limit"`
	CORS           CORSConfig           `yaml:"cors"`
	CircuitBreaker CircuitBreakerConfig `yaml:"circuit_breaker"`
	Cache          CacheConfig          `yaml:"cache"`
	LoadBalancing  LoadBalancingConfig  `yaml:"load_balancing"`
	Validation     ValidationConfig     `yaml:"validation"`
	Logging        LoggingConfig        `yaml:"logging"`
	Metrics        MetricsConfig        `yaml:"metrics"`
	Health         HealthConfig         `yaml:"health"`
	Versioning     VersioningConfig     `yaml:"versioning"`
	TLS            TLSConfig            `yaml:"tls"`
	Transformation TransformationConfig `yaml:"transformation"`
	WebSocket      WebSocketConfig      `yaml:"websocket"`
	Retry          RetryConfig          `yaml:"retry"`
	Timeout        TimeoutConfig        `yaml:"timeout"`
	Features       FeaturesConfig       `yaml:"features"`
}

type ServerConfig struct {
	Host           string        `yaml:"host"`
	Port           int           `yaml:"port"`
	Mode           string        `yaml:"mode"`
	ReadTimeout    time.Duration `yaml:"read_timeout"`
	WriteTimeout   time.Duration `yaml:"write_timeout"`
	IdleTimeout    time.Duration `yaml:"idle_timeout"`
	MaxHeaderBytes int           `yaml:"max_header_bytes"`
}

type ServiceConfig struct {
	URL        string        `yaml:"url"`
	Timeout    time.Duration `yaml:"timeout"`
	MaxRetries int           `yaml:"max_retries"`
}

type ServicesConfig struct {
	MarketData      ServiceConfig `yaml:"market_data"`
	OrderExecution  ServiceConfig `yaml:"order_execution"`
	StrategyEngine  ServiceConfig `yaml:"strategy_engine"`
	AccountMonitor  ServiceConfig `yaml:"account_monitor"`
	DashboardServer ServiceConfig `yaml:"dashboard_server"`
	RiskManager     ServiceConfig `yaml:"risk_manager"`
	Configuration   ServiceConfig `yaml:"configuration"`
}

type APIKey struct {
	Key  string `yaml:"key"`
	Role string `yaml:"role"`
}

type AuthConfig struct {
	Enabled            bool          `yaml:"enabled"`
	JWTSecret          string        `yaml:"jwt_secret"`
	JWTExpiry          time.Duration `yaml:"jwt_expiry"`
	RefreshTokenExpiry time.Duration `yaml:"refresh_token_expiry"`
	APIKeys            []APIKey      `yaml:"api_keys"`
}

type EndpointRateLimit struct {
	RequestsPerSecond int `yaml:"requests_per_second"`
	Burst             int `yaml:"burst"`
}

type GlobalRateLimit struct {
	RequestsPerSecond int `yaml:"requests_per_second"`
	Burst             int `yaml:"burst"`
}

type PerIPRateLimit struct {
	RequestsPerMinute int `yaml:"requests_per_minute"`
	Burst             int `yaml:"burst"`
}

type RateLimitConfig struct {
	Enabled   bool                         `yaml:"enabled"`
	Global    GlobalRateLimit              `yaml:"global"`
	Endpoints map[string]EndpointRateLimit `yaml:"endpoints"`
	PerIP     PerIPRateLimit               `yaml:"per_ip"`
}

type CORSConfig struct {
	Enabled          bool     `yaml:"enabled"`
	AllowedOrigins   []string `yaml:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers"`
	ExposedHeaders   []string `yaml:"expose_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
	MaxAge           int      `yaml:"max_age"`
}

type ServiceBreakerConfig struct {
	MaxRequests uint32        `yaml:"max_requests"`
	Interval    time.Duration `yaml:"interval"`
	Timeout     time.Duration `yaml:"timeout"`
}

type CircuitBreakerConfig struct {
	Enabled     bool                            `yaml:"enabled"`
	MaxRequests uint32                          `yaml:"max_requests"`
	Interval    time.Duration                   `yaml:"interval"`
	Timeout     time.Duration                   `yaml:"timeout"`
	Services    map[string]ServiceBreakerConfig `yaml:"services"`
}

type CacheRule struct {
	TTL time.Duration `yaml:"ttl"`
}

type CacheConfig struct {
	Enabled    bool                 `yaml:"enabled"`
	RedisURL   string               `yaml:"redis_url"`
	DefaultTTL time.Duration        `yaml:"default_ttl"`
	Rules      map[string]CacheRule `yaml:"rules"`
}

type LoadBalancingConfig struct {
	Enabled             bool          `yaml:"enabled"`
	Algorithm           string        `yaml:"algorithm"`
	HealthCheckInterval time.Duration `yaml:"health_check_interval"`
}

type ValidationConfig struct {
	Enabled        bool              `yaml:"enabled"`
	MaxRequestSize int64             `yaml:"max_request_size"`
	Schemas        map[string]string `yaml:"schemas"`
}

type LoggingConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	Output     string `yaml:"output"`
	FilePath   string `yaml:"file_path"`
	MaxSize    int    `yaml:"max_size"`
	MaxAge     int    `yaml:"max_age"`
	MaxBackups int    `yaml:"max_backups"`
	Compress   bool   `yaml:"compress"`
}

type MetricsConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Path      string `yaml:"path"`
	Namespace string `yaml:"namespace"`
}

type HealthConfig struct {
	Enabled        bool          `yaml:"enabled"`
	Path           string        `yaml:"path"`
	CheckServices  bool          `yaml:"check_services"`
	ServiceTimeout time.Duration `yaml:"service_timeout"`
}

type VersioningConfig struct {
	DefaultVersion     string   `yaml:"default_version"`
	SupportedVersions  []string `yaml:"supported_versions"`
	HeaderName         string   `yaml:"header_name"`
}

type TLSConfig struct {
	Enabled    bool   `yaml:"enabled"`
	CertFile   string `yaml:"cert_file"`
	KeyFile    string `yaml:"key_file"`
	MinVersion string `yaml:"min_version"`
}

type TransformationConfig struct {
	Enabled               bool              `yaml:"enabled"`
	RequestHeaders        map[string]string `yaml:"request_headers"`
	ResponseHeadersRemove []string          `yaml:"response_headers_remove"`
}

type WebSocketConfig struct {
	Enabled          bool          `yaml:"enabled"`
	MaxConnections   int           `yaml:"max_connections"`
	ReadBufferSize   int           `yaml:"read_buffer_size"`
	WriteBufferSize  int           `yaml:"write_buffer_size"`
	HandshakeTimeout time.Duration `yaml:"handshake_timeout"`
	PingInterval     time.Duration `yaml:"ping_interval"`
	PongTimeout      time.Duration `yaml:"pong_timeout"`
}

type RetryConfig struct {
	Enabled              bool          `yaml:"enabled"`
	MaxAttempts          int           `yaml:"max_attempts"`
	BackoffMultiplier    int           `yaml:"backoff_multiplier"`
	InitialInterval      time.Duration `yaml:"initial_interval"`
	MaxInterval          time.Duration `yaml:"max_interval"`
	RetryableStatusCodes []int         `yaml:"retryable_status_codes"`
}

type TimeoutConfig struct {
	Default   time.Duration            `yaml:"default"`
	Max       time.Duration            `yaml:"max"`
	Endpoints map[string]time.Duration `yaml:"endpoints"`
}

type FeaturesConfig struct {
	EnableTracing      bool `yaml:"enable_tracing"`
	EnableCompression  bool `yaml:"enable_compression"`
	EnableRequestID    bool `yaml:"enable_request_id"`
	EnableAccessLog    bool `yaml:"enable_access_log"`
	EnableErrorDetails bool `yaml:"enable_error_details"`
}

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Expand environment variables
	expanded := os.ExpandEnv(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// Validate performs validation on the configuration
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Auth.Enabled && c.Auth.JWTSecret == "" {
		return fmt.Errorf("JWT secret is required when auth is enabled")
	}

	if c.Cache.Enabled && c.Cache.RedisURL == "" {
		return fmt.Errorf("Redis URL is required when cache is enabled")
	}

	return nil
}

// GetServiceURL returns the URL for a given service
func (c *Config) GetServiceURL(service string) string {
	switch service {
	case "market_data":
		return c.Services.MarketData.URL
	case "order_execution":
		return c.Services.OrderExecution.URL
	case "strategy_engine":
		return c.Services.StrategyEngine.URL
	case "account_monitor":
		return c.Services.AccountMonitor.URL
	case "dashboard_server":
		return c.Services.DashboardServer.URL
	case "risk_manager":
		return c.Services.RiskManager.URL
	case "configuration":
		return c.Services.Configuration.URL
	default:
		return ""
	}
}
