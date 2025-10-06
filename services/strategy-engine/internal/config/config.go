package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Redis      RedisConfig      `yaml:"redis"`
	NATS       NATSConfig       `yaml:"nats"`
	GRPC       GRPCConfig       `yaml:"grpc"`
	Engine     EngineConfig     `yaml:"engine"`
	Strategies StrategiesConfig `yaml:"strategies"`
	Risk       RiskConfig       `yaml:"risk"`
	Logging    LoggingConfig    `yaml:"logging"`
	Metrics    MetricsConfig    `yaml:"metrics"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host           string        `yaml:"host"`
	Port           int           `yaml:"port"`
	Mode           string        `yaml:"mode"` // development, production
	ReadTimeout    time.Duration `yaml:"readTimeout"`
	WriteTimeout   time.Duration `yaml:"writeTimeout"`
	IdleTimeout    time.Duration `yaml:"idleTimeout"`
	MaxHeaderBytes int           `yaml:"maxHeaderBytes"`
	APIKey         string        `yaml:"apiKey"`         // Optional API key for authentication
	EnableAuth     bool          `yaml:"enableAuth"`     // Enable API key authentication
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host               string   `yaml:"host"`
	Port               int      `yaml:"port"`
	Password           string   `yaml:"password"`
	DB                 int      `yaml:"db"`
	PoolSize           int      `yaml:"poolSize"`
	MinIdleConns       int      `yaml:"minIdleConns"`
	MarketDataChannels []string `yaml:"marketDataChannels"`
}

// NATSConfig holds NATS configuration
type NATSConfig struct {
	URL             string `yaml:"url"`
	MaxReconnects   int    `yaml:"maxReconnects"`
	ReconnectWait   int    `yaml:"reconnectWait"` // seconds
	FillSubject     string `yaml:"fillSubject"`
	PositionSubject string `yaml:"positionSubject"`
}

// GRPCConfig holds gRPC client configuration
type GRPCConfig struct {
	OrderExecutionAddr string        `yaml:"orderExecutionAddr"`
	Timeout            time.Duration `yaml:"timeout"`
	MaxRetries         int           `yaml:"maxRetries"`
}

// EngineConfig holds engine-specific configuration
type EngineConfig struct {
	Mode              string        `yaml:"mode"` // live, simulation, observation
	SignalBufferSize  int           `yaml:"signalBufferSize"`
	MaxConcurrent     int           `yaml:"maxConcurrent"`
	ProcessingTimeout time.Duration `yaml:"processingTimeout"`
	PluginsDir        string        `yaml:"pluginsDir"`
	HotReload         bool          `yaml:"hotReload"`
	ReloadInterval    time.Duration `yaml:"reloadInterval"`
}

// StrategiesConfig holds strategies configuration
type StrategiesConfig struct {
	Enabled     []string               `yaml:"enabled"`
	Configs     map[string]interface{} `yaml:"configs"`
	PythonPath  string                 `yaml:"pythonPath"`
	PythonVenv  string                 `yaml:"pythonVenv"`
}

// RiskConfig holds risk management configuration
type RiskConfig struct {
	Enabled             bool    `yaml:"enabled"`
	MaxPositionSize     float64 `yaml:"maxPositionSize"`
	MaxOrderValue       float64 `yaml:"maxOrderValue"`
	MaxDailyLoss        float64 `yaml:"maxDailyLoss"`
	MaxDrawdown         float64 `yaml:"maxDrawdown"`
	MinAccountBalance   float64 `yaml:"minAccountBalance"`
	AllowedSymbols      []string `yaml:"allowedSymbols"`
	BlockedSymbols      []string `yaml:"blockedSymbols"`
	MaxOrdersPerSecond  int     `yaml:"maxOrdersPerSecond"`
	MaxOrdersPerMinute  int     `yaml:"maxOrdersPerMinute"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `yaml:"level"` // debug, info, warn, error
	Format     string `yaml:"format"` // json, console
	Output     string `yaml:"output"` // stdout, stderr, file
	File       string `yaml:"file"`
	MaxSize    int    `yaml:"maxSize"`    // megabytes
	MaxBackups int    `yaml:"maxBackups"`
	MaxAge     int    `yaml:"maxAge"`     // days
	Compress   bool   `yaml:"compress"`
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Port      int    `yaml:"port"`
	Path      string `yaml:"path"`
	Namespace string `yaml:"namespace"`
}

// Load reads configuration from file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Apply defaults
	applyDefaults(cfg)

	return cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 9092
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 10 * time.Second
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 10 * time.Second
	}
	if cfg.Server.IdleTimeout == 0 {
		cfg.Server.IdleTimeout = 60 * time.Second
	}
	if cfg.Server.MaxHeaderBytes == 0 {
		cfg.Server.MaxHeaderBytes = 1 << 20 // 1MB
	}

	if cfg.Redis.PoolSize == 0 {
		cfg.Redis.PoolSize = 10
	}
	if cfg.Redis.MinIdleConns == 0 {
		cfg.Redis.MinIdleConns = 5
	}

	if cfg.NATS.MaxReconnects == 0 {
		cfg.NATS.MaxReconnects = 10
	}
	if cfg.NATS.ReconnectWait == 0 {
		cfg.NATS.ReconnectWait = 2
	}
	if cfg.NATS.FillSubject == "" {
		cfg.NATS.FillSubject = "trading.fills"
	}
	if cfg.NATS.PositionSubject == "" {
		cfg.NATS.PositionSubject = "trading.positions"
	}

	if cfg.GRPC.Timeout == 0 {
		cfg.GRPC.Timeout = 5 * time.Second
	}
	if cfg.GRPC.MaxRetries == 0 {
		cfg.GRPC.MaxRetries = 3
	}

	if cfg.Engine.Mode == "" {
		cfg.Engine.Mode = "simulation"
	}
	if cfg.Engine.SignalBufferSize == 0 {
		cfg.Engine.SignalBufferSize = 1000
	}
	if cfg.Engine.MaxConcurrent == 0 {
		cfg.Engine.MaxConcurrent = 10
	}
	if cfg.Engine.ProcessingTimeout == 0 {
		cfg.Engine.ProcessingTimeout = 500 * time.Microsecond
	}
	if cfg.Engine.PluginsDir == "" {
		cfg.Engine.PluginsDir = "./plugins"
	}
	if cfg.Engine.ReloadInterval == 0 {
		cfg.Engine.ReloadInterval = 30 * time.Second
	}

	if cfg.Strategies.PythonPath == "" {
		cfg.Strategies.PythonPath = "/usr/bin/python3"
	}

	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = "json"
	}
	if cfg.Logging.Output == "" {
		cfg.Logging.Output = "stdout"
	}

	if cfg.Metrics.Path == "" {
		cfg.Metrics.Path = "/metrics"
	}
	if cfg.Metrics.Namespace == "" {
		cfg.Metrics.Namespace = "strategy_engine"
	}
}
