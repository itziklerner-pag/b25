package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Database  DatabaseConfig  `yaml:"database"`
	Redis     RedisConfig     `yaml:"redis"`
	Kafka     KafkaConfig     `yaml:"kafka"`
	Analytics AnalyticsConfig `yaml:"analytics"`
	Metrics   MetricsConfig   `yaml:"metrics"`
	Logging   LoggingConfig   `yaml:"logging"`
	Health    HealthConfig    `yaml:"health"`
	CustomEvents CustomEventsConfig `yaml:"custom_events"`
	Realtime  RealtimeConfig  `yaml:"realtime"`
	Security  SecurityConfig  `yaml:"security"`
}

type ServerConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type DatabaseConfig struct {
	Host               string        `yaml:"host"`
	Port               int           `yaml:"port"`
	Database           string        `yaml:"database"`
	User               string        `yaml:"user"`
	Password           string        `yaml:"password"`
	SSLMode            string        `yaml:"ssl_mode"`
	MaxConnections     int           `yaml:"max_connections"`
	MaxIdleConnections int           `yaml:"max_idle_connections"`
	ConnectionLifetime time.Duration `yaml:"connection_lifetime"`
}

type RedisConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Password     string `yaml:"password"`
	DB           int    `yaml:"db"`
	PoolSize     int    `yaml:"pool_size"`
	MinIdleConns int    `yaml:"min_idle_conns"`
}

type KafkaConfig struct {
	Brokers          []string      `yaml:"brokers"`
	ConsumerGroup    string        `yaml:"consumer_group"`
	Topics           []string      `yaml:"topics"`
	EnableAutoCommit bool          `yaml:"enable_auto_commit"`
	SessionTimeout   time.Duration `yaml:"session_timeout"`
}

type AnalyticsConfig struct {
	Ingestion   IngestionConfig   `yaml:"ingestion"`
	Aggregation AggregationConfig `yaml:"aggregation"`
	Retention   RetentionConfig   `yaml:"retention"`
	Query       QueryConfig       `yaml:"query"`
}

type IngestionConfig struct {
	BatchSize    int           `yaml:"batch_size"`
	BatchTimeout time.Duration `yaml:"batch_timeout"`
	Workers      int           `yaml:"workers"`
	BufferSize   int           `yaml:"buffer_size"`
}

type AggregationConfig struct {
	Intervals []string `yaml:"intervals"`
	Workers   int      `yaml:"workers"`
}

type RetentionConfig struct {
	RawEvents        time.Duration `yaml:"raw_events"`
	MinuteAggregates time.Duration `yaml:"minute_aggregates"`
	HourAggregates   time.Duration `yaml:"hour_aggregates"`
	DailyAggregates  time.Duration `yaml:"daily_aggregates"`
}

type QueryConfig struct {
	MaxResults   int           `yaml:"max_results"`
	DefaultLimit int           `yaml:"default_limit"`
	Timeout      time.Duration `yaml:"timeout"`
	CacheTTL     time.Duration `yaml:"cache_ttl"`
}

type MetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Port    int    `yaml:"port"`
	Path    string `yaml:"path"`
}

type LoggingConfig struct {
	Level    string `yaml:"level"`
	Format   string `yaml:"format"`
	Output   string `yaml:"output"`
	FilePath string `yaml:"file_path"`
}

type HealthConfig struct {
	Port int    `yaml:"port"`
	Path string `yaml:"path"`
}

type CustomEventsConfig struct {
	Enabled         bool `yaml:"enabled"`
	MaxProperties   int  `yaml:"max_properties"`
	MaxPropertySize int  `yaml:"max_property_size"`
}

type RealtimeConfig struct {
	Enabled        bool          `yaml:"enabled"`
	WebsocketPort  int           `yaml:"websocket_port"`
	MaxConnections int           `yaml:"max_connections"`
	UpdateInterval time.Duration `yaml:"update_interval"`
}

type SecurityConfig struct {
	APIKeyEnabled bool            `yaml:"api_key_enabled"`
	RateLimit     RateLimitConfig `yaml:"rate_limit"`
	CORS          CORSConfig      `yaml:"cors"`
}

type RateLimitConfig struct {
	Enabled            bool `yaml:"enabled"`
	RequestsPerMinute  int  `yaml:"requests_per_minute"`
	Burst              int  `yaml:"burst"`
}

type CORSConfig struct {
	Enabled        bool     `yaml:"enabled"`
	AllowedOrigins []string `yaml:"allowed_origins"`
	AllowedMethods []string `yaml:"allowed_methods"`
	AllowedHeaders []string `yaml:"allowed_headers"`
}

// Load loads configuration from a YAML file
func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Override with environment variables if present
	config.overrideFromEnv()

	return &config, nil
}

// overrideFromEnv overrides config values with environment variables
func (c *Config) overrideFromEnv() {
	if val := os.Getenv("SERVER_PORT"); val != "" {
		fmt.Sscanf(val, "%d", &c.Server.Port)
	}
	if val := os.Getenv("DB_HOST"); val != "" {
		c.Database.Host = val
	}
	if val := os.Getenv("DB_PORT"); val != "" {
		fmt.Sscanf(val, "%d", &c.Database.Port)
	}
	if val := os.Getenv("DB_NAME"); val != "" {
		c.Database.Database = val
	}
	if val := os.Getenv("DB_USER"); val != "" {
		c.Database.User = val
	}
	if val := os.Getenv("DB_PASSWORD"); val != "" {
		c.Database.Password = val
	}
	if val := os.Getenv("REDIS_HOST"); val != "" {
		c.Redis.Host = val
	}
	if val := os.Getenv("REDIS_PORT"); val != "" {
		fmt.Sscanf(val, "%d", &c.Redis.Port)
	}
	if val := os.Getenv("KAFKA_BROKERS"); val != "" {
		c.Kafka.Brokers = []string{val}
	}
	if val := os.Getenv("LOG_LEVEL"); val != "" {
		c.Logging.Level = val
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.Database == "" {
		return fmt.Errorf("database name is required")
	}
	if len(c.Kafka.Brokers) == 0 {
		return fmt.Errorf("at least one Kafka broker is required")
	}
	return nil
}
