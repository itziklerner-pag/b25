package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the Risk Manager Service
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	NATS     NATSConfig
	GRPC     GRPCConfig
	Risk     RiskConfig
	Logging  LoggingConfig
	Metrics  MetricsConfig
}

type ServerConfig struct {
	Port            int
	Mode            string // "development" or "production"
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Host       string
	Port       int
	Password   string
	DB         int
	MaxRetries int
	PoolSize   int
}

type NATSConfig struct {
	URL             string
	MaxReconnect    int
	ReconnectWait   time.Duration
	AlertSubject    string
	EmergencyTopic  string
}

type GRPCConfig struct {
	Port                int
	MaxConnectionIdle   time.Duration
	MaxConnectionAge    time.Duration
	KeepAliveInterval   time.Duration
	KeepAliveTimeout    time.Duration
}

type RiskConfig struct {
	MonitorInterval      time.Duration
	CacheTTL             time.Duration
	PolicyCacheTTL       time.Duration
	MaxLeverage          float64
	MaxDrawdownPercent   float64
	EmergencyThreshold   float64
	AlertWindow          time.Duration
	AccountMonitorURL    string
	MarketDataRedisDB    int
}

type LoggingConfig struct {
	Level  string
	Format string // "json" or "console"
}

type MetricsConfig struct {
	Enabled bool
	Port    int
}

// Load configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Config file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
		v.AddConfigPath("/etc/risk-manager")
	}

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Environment variables
	v.SetEnvPrefix("RISK")
	v.AutomaticEnv()

	// Unmarshal
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	// Server
	v.SetDefault("server.port", 9095)
	v.SetDefault("server.mode", "development")
	v.SetDefault("server.readTimeout", 10*time.Second)
	v.SetDefault("server.writeTimeout", 10*time.Second)
	v.SetDefault("server.shutdownTimeout", 15*time.Second)

	// Database
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "postgres")
	v.SetDefault("database.database", "risk_manager")
	v.SetDefault("database.sslMode", "disable")
	v.SetDefault("database.maxOpenConns", 25)
	v.SetDefault("database.maxIdleConns", 5)
	v.SetDefault("database.connMaxLifetime", 5*time.Minute)

	// Redis
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.maxRetries", 3)
	v.SetDefault("redis.poolSize", 10)

	// NATS
	v.SetDefault("nats.url", "nats://localhost:4222")
	v.SetDefault("nats.maxReconnect", 10)
	v.SetDefault("nats.reconnectWait", 2*time.Second)
	v.SetDefault("nats.alertSubject", "risk.alerts")
	v.SetDefault("nats.emergencyTopic", "risk.emergency")

	// gRPC
	v.SetDefault("grpc.port", 50051)
	v.SetDefault("grpc.maxConnectionIdle", 5*time.Minute)
	v.SetDefault("grpc.maxConnectionAge", 30*time.Minute)
	v.SetDefault("grpc.keepAliveInterval", 30*time.Second)
	v.SetDefault("grpc.keepAliveTimeout", 10*time.Second)

	// Risk
	v.SetDefault("risk.monitorInterval", 1*time.Second)
	v.SetDefault("risk.cacheTTL", 100*time.Millisecond)
	v.SetDefault("risk.policyCacheTTL", 1*time.Second)
	v.SetDefault("risk.maxLeverage", 10.0)
	v.SetDefault("risk.maxDrawdownPercent", 0.20)
	v.SetDefault("risk.emergencyThreshold", 0.25)
	v.SetDefault("risk.alertWindow", 5*time.Minute)
	v.SetDefault("risk.accountMonitorURL", "localhost:50053")
	v.SetDefault("risk.marketDataRedisDB", 1)

	// Logging
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")

	// Metrics
	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.port", 9095)
}

// GetDSN returns database connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}

// GetRedisAddr returns Redis address
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetServerAddr returns HTTP server address
func (c *ServerConfig) GetServerAddr() string {
	return fmt.Sprintf(":%d", c.Port)
}

// GetGRPCAddr returns gRPC server address
func (c *GRPCConfig) GetGRPCAddr() string {
	return fmt.Sprintf(":%d", c.Port)
}
