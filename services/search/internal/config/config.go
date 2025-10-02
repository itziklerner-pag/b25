package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server         ServerConfig         `mapstructure:"server"`
	Elasticsearch  ElasticsearchConfig  `mapstructure:"elasticsearch"`
	Redis          RedisConfig          `mapstructure:"redis"`
	NATS           NATSConfig           `mapstructure:"nats"`
	Search         SearchConfig         `mapstructure:"search"`
	Indexer        IndexerConfig        `mapstructure:"indexer"`
	Analytics      AnalyticsConfig      `mapstructure:"analytics"`
	Logging        LoggingConfig        `mapstructure:"logging"`
	Metrics        MetricsConfig        `mapstructure:"metrics"`
	Health         HealthConfig         `mapstructure:"health"`
	RateLimit      RateLimitConfig      `mapstructure:"rate_limit"`
	CORS           CORSConfig           `mapstructure:"cors"`
}

// ServerConfig contains HTTP server settings
type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// ElasticsearchConfig contains Elasticsearch settings
type ElasticsearchConfig struct {
	Addresses          []string              `mapstructure:"addresses"`
	Username           string                `mapstructure:"username"`
	Password           string                `mapstructure:"password"`
	APIKey             string                `mapstructure:"api_key"`
	MaxRetries         int                   `mapstructure:"max_retries"`
	RetryBackoff       time.Duration         `mapstructure:"retry_backoff"`
	EnableDebugLogger  bool                  `mapstructure:"enable_debug_logger"`
	Indices            map[string]IndexConfig `mapstructure:"indices"`
}

// IndexConfig contains index settings
type IndexConfig struct {
	Name     string `mapstructure:"name"`
	Shards   int    `mapstructure:"shards"`
	Replicas int    `mapstructure:"replicas"`
}

// RedisConfig contains Redis settings
type RedisConfig struct {
	Address      string        `mapstructure:"address"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	PoolSize     int           `mapstructure:"pool_size"`
	CacheTTL     time.Duration `mapstructure:"cache_ttl"`
	AnalyticsTTL time.Duration `mapstructure:"analytics_ttl"`
}

// NATSConfig contains NATS settings
type NATSConfig struct {
	URL      string          `mapstructure:"url"`
	Subjects SubjectsConfig  `mapstructure:"subjects"`
}

// SubjectsConfig contains NATS subject patterns
type SubjectsConfig struct {
	Trades     string `mapstructure:"trades"`
	Orders     string `mapstructure:"orders"`
	Strategies string `mapstructure:"strategies"`
	MarketData string `mapstructure:"market_data"`
	Logs       string `mapstructure:"logs"`
}

// SearchConfig contains search-specific settings
type SearchConfig struct {
	MaxResults      int                 `mapstructure:"max_results"`
	DefaultPageSize int                 `mapstructure:"default_page_size"`
	MaxPageSize     int                 `mapstructure:"max_page_size"`
	Timeout         time.Duration       `mapstructure:"timeout"`
	MinScore        float64             `mapstructure:"min_score"`
	Autocomplete    AutocompleteConfig  `mapstructure:"autocomplete"`
	Facets          FacetsConfig        `mapstructure:"facets"`
	Highlight       HighlightConfig     `mapstructure:"highlight"`
}

// AutocompleteConfig contains autocomplete settings
type AutocompleteConfig struct {
	MinChars       int `mapstructure:"min_chars"`
	MaxSuggestions int `mapstructure:"max_suggestions"`
}

// FacetsConfig contains facet settings
type FacetsConfig struct {
	MaxBuckets int `mapstructure:"max_buckets"`
}

// HighlightConfig contains highlighting settings
type HighlightConfig struct {
	Enabled           bool   `mapstructure:"enabled"`
	PreTag            string `mapstructure:"pre_tag"`
	PostTag           string `mapstructure:"post_tag"`
	FragmentSize      int    `mapstructure:"fragment_size"`
	NumberOfFragments int    `mapstructure:"number_of_fragments"`
}

// IndexerConfig contains indexer settings
type IndexerConfig struct {
	BatchSize     int           `mapstructure:"batch_size"`
	FlushInterval time.Duration `mapstructure:"flush_interval"`
	Workers       int           `mapstructure:"workers"`
	MaxRetries    int           `mapstructure:"max_retries"`
	RetryDelay    time.Duration `mapstructure:"retry_delay"`
	QueueBuffer   int           `mapstructure:"queue_buffer"`
}

// AnalyticsConfig contains analytics settings
type AnalyticsConfig struct {
	Enabled        bool `mapstructure:"enabled"`
	TrackSearches  bool `mapstructure:"track_searches"`
	TrackClicks    bool `mapstructure:"track_clicks"`
	RetentionDays  int  `mapstructure:"retention_days"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level    string `mapstructure:"level"`
	Format   string `mapstructure:"format"`
	Output   string `mapstructure:"output"`
	FilePath string `mapstructure:"file_path"`
}

// MetricsConfig contains metrics settings
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
	Path    string `mapstructure:"path"`
}

// HealthConfig contains health check settings
type HealthConfig struct {
	Path          string `mapstructure:"path"`
	ReadinessPath string `mapstructure:"readiness_path"`
}

// RateLimitConfig contains rate limiting settings
type RateLimitConfig struct {
	Enabled            bool    `mapstructure:"enabled"`
	RequestsPerSecond  float64 `mapstructure:"requests_per_second"`
	Burst              int     `mapstructure:"burst"`
}

// CORSConfig contains CORS settings
type CORSConfig struct {
	Enabled        bool     `mapstructure:"enabled"`
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	AllowedMethods []string `mapstructure:"allowed_methods"`
	AllowedHeaders []string `mapstructure:"allowed_headers"`
	MaxAge         int      `mapstructure:"max_age"`
}

// Load loads configuration from file
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set config file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
		v.AddConfigPath("/etc/search-service")
	}

	// Set defaults
	setDefaults(v)

	// Read from environment variables
	v.SetEnvPrefix("SEARCH")
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
		// Config file not found, use defaults
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 9097)
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "30s")
	v.SetDefault("server.shutdown_timeout", "10s")

	// Elasticsearch defaults
	v.SetDefault("elasticsearch.addresses", []string{"http://localhost:9200"})
	v.SetDefault("elasticsearch.max_retries", 3)
	v.SetDefault("elasticsearch.retry_backoff", "1s")
	v.SetDefault("elasticsearch.enable_debug_logger", false)

	// Redis defaults
	v.SetDefault("redis.address", "localhost:6379")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 10)
	v.SetDefault("redis.cache_ttl", "300s")
	v.SetDefault("redis.analytics_ttl", "3600s")

	// NATS defaults
	v.SetDefault("nats.url", "nats://localhost:4222")
	v.SetDefault("nats.subjects.trades", "trades.>")
	v.SetDefault("nats.subjects.orders", "orders.>")
	v.SetDefault("nats.subjects.strategies", "strategies.>")
	v.SetDefault("nats.subjects.market_data", "market.>")
	v.SetDefault("nats.subjects.logs", "logs.>")

	// Search defaults
	v.SetDefault("search.max_results", 10000)
	v.SetDefault("search.default_page_size", 50)
	v.SetDefault("search.max_page_size", 1000)
	v.SetDefault("search.timeout", "5s")
	v.SetDefault("search.min_score", 0.1)
	v.SetDefault("search.autocomplete.min_chars", 2)
	v.SetDefault("search.autocomplete.max_suggestions", 10)
	v.SetDefault("search.facets.max_buckets", 100)
	v.SetDefault("search.highlight.enabled", true)
	v.SetDefault("search.highlight.pre_tag", "<em>")
	v.SetDefault("search.highlight.post_tag", "</em>")
	v.SetDefault("search.highlight.fragment_size", 150)
	v.SetDefault("search.highlight.number_of_fragments", 3)

	// Indexer defaults
	v.SetDefault("indexer.batch_size", 1000)
	v.SetDefault("indexer.flush_interval", "5s")
	v.SetDefault("indexer.workers", 4)
	v.SetDefault("indexer.max_retries", 3)
	v.SetDefault("indexer.retry_delay", "1s")
	v.SetDefault("indexer.queue_buffer", 10000)

	// Analytics defaults
	v.SetDefault("analytics.enabled", true)
	v.SetDefault("analytics.track_searches", true)
	v.SetDefault("analytics.track_clicks", true)
	v.SetDefault("analytics.retention_days", 90)

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.output", "stdout")

	// Metrics defaults
	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.port", 9098)
	v.SetDefault("metrics.path", "/metrics")

	// Health defaults
	v.SetDefault("health.path", "/health")
	v.SetDefault("health.readiness_path", "/ready")

	// Rate limit defaults
	v.SetDefault("rate_limit.enabled", true)
	v.SetDefault("rate_limit.requests_per_second", 100)
	v.SetDefault("rate_limit.burst", 200)

	// CORS defaults
	v.SetDefault("cors.enabled", true)
	v.SetDefault("cors.allowed_origins", []string{"http://localhost:3000"})
	v.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE"})
	v.SetDefault("cors.allowed_headers", []string{"Content-Type", "Authorization"})
	v.SetDefault("cors.max_age", 3600)
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if len(c.Elasticsearch.Addresses) == 0 {
		return fmt.Errorf("at least one Elasticsearch address is required")
	}

	if c.Search.MaxPageSize > c.Search.MaxResults {
		return fmt.Errorf("max_page_size cannot exceed max_results")
	}

	if c.Indexer.Workers <= 0 {
		return fmt.Errorf("indexer workers must be positive")
	}

	return nil
}
