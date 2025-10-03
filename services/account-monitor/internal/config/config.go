package config

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
)

type Config struct {
	Service        ServiceConfig        `mapstructure:"service"`
	GRPC           GRPCConfig           `mapstructure:"grpc"`
	HTTP           HTTPConfig           `mapstructure:"http"`
	Metrics        MetricsConfig        `mapstructure:"metrics"`
	Exchange       ExchangeConfig       `mapstructure:"exchange"`
	Database       DatabaseConfig       `mapstructure:"database"`
	Reconciliation ReconciliationConfig `mapstructure:"reconciliation"`
	Alerts         AlertsConfig         `mapstructure:"alerts"`
	PubSub         PubSubConfig         `mapstructure:"pubsub"`
	Logging        LoggingConfig        `mapstructure:"logging"`
}

type ServiceConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
}

type GRPCConfig struct {
	Port           int `mapstructure:"port"`
	MaxConnections int `mapstructure:"max_connections"`
}

type HTTPConfig struct {
	Port             int  `mapstructure:"port"`
	DashboardEnabled bool `mapstructure:"dashboard_enabled"`
}

type MetricsConfig struct {
	Port int    `mapstructure:"port"`
	Path string `mapstructure:"path"`
}

type ExchangeConfig struct {
	Name         string          `mapstructure:"name"`
	Testnet      bool            `mapstructure:"testnet"`
	APIKey       string          `mapstructure:"api_key"`
	SecretKey    string          `mapstructure:"secret_key"`
	APIKeyEnv    string          `mapstructure:"api_key_env"`
	SecretKeyEnv string          `mapstructure:"secret_key_env"`
	WebSocket    WebSocketConfig `mapstructure:"websocket"`
}

type WebSocketConfig struct {
	ReconnectInterval time.Duration `mapstructure:"reconnect_interval"`
	PingInterval      time.Duration `mapstructure:"ping_interval"`
	Timeout           time.Duration `mapstructure:"timeout"`
}

type DatabaseConfig struct {
	Postgres PostgresConfig `mapstructure:"postgres"`
	Redis    RedisConfig    `mapstructure:"redis"`
}

type PostgresConfig struct {
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	Database       string `mapstructure:"database"`
	User           string `mapstructure:"user"`
	Password       string `mapstructure:"password"`
	PasswordEnv    string `mapstructure:"password_env"`
	MaxConnections int    `mapstructure:"max_connections"`
	SSLMode        string `mapstructure:"ssl_mode"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	DB       int    `mapstructure:"db"`
	Password string `mapstructure:"password"`
}

type ReconciliationConfig struct {
	Enabled           bool            `mapstructure:"enabled"`
	Interval          time.Duration   `mapstructure:"interval"`
	BalanceTolerance  decimal.Decimal `mapstructure:"balance_tolerance"`
	PositionTolerance decimal.Decimal `mapstructure:"position_tolerance"`
}

type AlertsConfig struct {
	Enabled             bool                 `mapstructure:"enabled"`
	Thresholds          AlertThresholds      `mapstructure:"thresholds"`
	SuppressionDuration time.Duration        `mapstructure:"suppression_duration"`
}

type AlertThresholds struct {
	MinBalance        decimal.Decimal `mapstructure:"min_balance"`
	MaxDrawdownPct    decimal.Decimal `mapstructure:"max_drawdown_pct"`
	MaxMarginRatio    decimal.Decimal `mapstructure:"max_margin_ratio"`
	BalanceDriftPct   decimal.Decimal `mapstructure:"balance_drift_pct"`
	PositionDriftPct  decimal.Decimal `mapstructure:"position_drift_pct"`
}

type PubSubConfig struct {
	Provider string      `mapstructure:"provider"`
	NATS     NATSConfig  `mapstructure:"nats"`
	Topics   TopicsConfig `mapstructure:"topics"`
}

type NATSConfig struct {
	URL            string        `mapstructure:"url"`
	MaxReconnects  int           `mapstructure:"max_reconnects"`
	ReconnectWait  time.Duration `mapstructure:"reconnect_wait"`
}

type TopicsConfig struct {
	FillEvents string `mapstructure:"fill_events"`
	Alerts     string `mapstructure:"alerts"`
	PnLUpdates string `mapstructure:"pnl_updates"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// stringToDecimalHookFunc is a custom decode hook that converts strings to decimal.Decimal
func stringToDecimalHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{},
	) (interface{}, error) {
		// Check if we're converting to decimal.Decimal
		if t != reflect.TypeOf(decimal.Decimal{}) {
			return data, nil
		}

		// Check if the source is a string
		if f.Kind() != reflect.String {
			return data, nil
		}

		// Convert string to decimal.Decimal
		strVal := data.(string)
		decVal, err := decimal.NewFromString(strVal)
		if err != nil {
			return nil, fmt.Errorf("failed to parse decimal from string '%s': %w", strVal, err)
		}

		return decVal, nil
	}
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found, will use defaults and env vars
	}

	var cfg Config

	// Create decoder with custom decode hook for decimal.Decimal
	decoderConfig := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			stringToDecimalHookFunc(),
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
		Result:           &cfg,
		WeaklyTypedInput: true,
	}

	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create decoder: %w", err)
	}

	if err := decoder.Decode(viper.AllSettings()); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Load secrets from environment variables
	if cfg.Exchange.APIKeyEnv != "" {
		if key := os.Getenv(cfg.Exchange.APIKeyEnv); key != "" {
			cfg.Exchange.APIKey = key
		}
	}
	if cfg.Exchange.SecretKeyEnv != "" {
		if secret := os.Getenv(cfg.Exchange.SecretKeyEnv); secret != "" {
			cfg.Exchange.SecretKey = secret
		}
	}
	if cfg.Database.Postgres.PasswordEnv != "" {
		if password := os.Getenv(cfg.Database.Postgres.PasswordEnv); password != "" {
			cfg.Database.Postgres.Password = password
		}
	}

	// Set defaults
	setDefaults(&cfg)

	return &cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.Service.Name == "" {
		cfg.Service.Name = "account-monitor"
	}
	if cfg.Service.Version == "" {
		cfg.Service.Version = "1.0.0"
	}
	if cfg.GRPC.Port == 0 {
		cfg.GRPC.Port = 50051
	}
	if cfg.GRPC.MaxConnections == 0 {
		cfg.GRPC.MaxConnections = 100
	}
	if cfg.HTTP.Port == 0 {
		cfg.HTTP.Port = 8080
	}
	if cfg.Metrics.Port == 0 {
		cfg.Metrics.Port = 9093
	}
	if cfg.Metrics.Path == "" {
		cfg.Metrics.Path = "/metrics"
	}
	if cfg.Exchange.WebSocket.ReconnectInterval == 0 {
		cfg.Exchange.WebSocket.ReconnectInterval = 5 * time.Second
	}
	if cfg.Exchange.WebSocket.PingInterval == 0 {
		cfg.Exchange.WebSocket.PingInterval = 30 * time.Second
	}
	if cfg.Exchange.WebSocket.Timeout == 0 {
		cfg.Exchange.WebSocket.Timeout = 60 * time.Second
	}
	if cfg.Database.Postgres.Port == 0 {
		cfg.Database.Postgres.Port = 5432
	}
	if cfg.Database.Postgres.MaxConnections == 0 {
		cfg.Database.Postgres.MaxConnections = 10
	}
	if cfg.Database.Postgres.SSLMode == "" {
		cfg.Database.Postgres.SSLMode = "disable"
	}
	if cfg.Database.Redis.Port == 0 {
		cfg.Database.Redis.Port = 6379
	}
	if cfg.Reconciliation.Interval == 0 {
		cfg.Reconciliation.Interval = 5 * time.Second
	}
	if cfg.Alerts.SuppressionDuration == 0 {
		cfg.Alerts.SuppressionDuration = 60 * time.Second
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
	if cfg.PubSub.NATS.MaxReconnects == 0 {
		cfg.PubSub.NATS.MaxReconnects = 10
	}
	if cfg.PubSub.NATS.ReconnectWait == 0 {
		cfg.PubSub.NATS.ReconnectWait = 2 * time.Second
	}
}
