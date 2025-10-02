package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server           ServerConfig           `mapstructure:"server"`
	Database         DatabaseConfig         `mapstructure:"database"`
	Redis            RedisConfig            `mapstructure:"redis"`
	Queue            QueueConfig            `mapstructure:"queue"`
	Email            EmailConfig            `mapstructure:"email"`
	SMS              SMSConfig              `mapstructure:"sms"`
	Push             PushConfig             `mapstructure:"push"`
	Preferences      PreferencesConfig      `mapstructure:"preferences"`
	RateLimit        RateLimitConfig        `mapstructure:"rate_limit"`
	Templates        TemplatesConfig        `mapstructure:"templates"`
	Logging          LoggingConfig          `mapstructure:"logging"`
	Metrics          MetricsConfig          `mapstructure:"metrics"`
	Health           HealthConfig           `mapstructure:"health"`
	ServiceDiscovery ServiceDiscoveryConfig `mapstructure:"service_discovery"`
	Subscriptions    SubscriptionsConfig    `mapstructure:"subscriptions"`
	AlertRules       AlertRulesConfig       `mapstructure:"alert_rules"`
	Services         ServicesConfig         `mapstructure:"services"`
}

type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Mode            string        `mapstructure:"mode"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Password   string `mapstructure:"password"`
	DB         int    `mapstructure:"db"`
	MaxRetries int    `mapstructure:"max_retries"`
	PoolSize   int    `mapstructure:"pool_size"`
}

type QueueConfig struct {
	RedisAddr       string        `mapstructure:"redis_addr"`
	Concurrency     int           `mapstructure:"concurrency"`
	MaxRetry        int           `mapstructure:"max_retry"`
	RetryDelayBase  time.Duration `mapstructure:"retry_delay_base"`
	RetentionDays   int           `mapstructure:"retention_days"`
}

type EmailConfig struct {
	Provider    string        `mapstructure:"provider"`
	FromAddress string        `mapstructure:"from_address"`
	FromName    string        `mapstructure:"from_name"`
	SendGrid    SendGridConfig `mapstructure:"sendgrid"`
	SMTP        SMTPConfig     `mapstructure:"smtp"`
}

type SendGridConfig struct {
	APIKey string `mapstructure:"api_key"`
}

type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	UseTLS   bool   `mapstructure:"use_tls"`
}

type SMSConfig struct {
	Provider string       `mapstructure:"provider"`
	Twilio   TwilioConfig `mapstructure:"twilio"`
}

type TwilioConfig struct {
	AccountSID string `mapstructure:"account_sid"`
	AuthToken  string `mapstructure:"auth_token"`
	FromNumber string `mapstructure:"from_number"`
}

type PushConfig struct {
	Provider string    `mapstructure:"provider"`
	FCM      FCMConfig `mapstructure:"fcm"`
}

type FCMConfig struct {
	CredentialsFile string `mapstructure:"credentials_file"`
	ProjectID       string `mapstructure:"project_id"`
}

type PreferencesConfig struct {
	DefaultChannels     []string `mapstructure:"default_channels"`
	OptInRequired       bool     `mapstructure:"opt_in_required"`
	QuietHoursStart     string   `mapstructure:"quiet_hours_start"`
	QuietHoursEnd       string   `mapstructure:"quiet_hours_end"`
	RespectQuietHours   bool     `mapstructure:"respect_quiet_hours"`
}

type RateLimitConfig struct {
	EmailPerHour      int `mapstructure:"email_per_hour"`
	SMSPerHour        int `mapstructure:"sms_per_hour"`
	PushPerHour       int `mapstructure:"push_per_hour"`
	GlobalPerSecond   int `mapstructure:"global_per_second"`
}

type TemplatesConfig struct {
	Directory    string `mapstructure:"directory"`
	AutoReload   bool   `mapstructure:"auto_reload"`
	CacheEnabled bool   `mapstructure:"cache_enabled"`
}

type LoggingConfig struct {
	Level       string `mapstructure:"level"`
	Format      string `mapstructure:"format"`
	Output      string `mapstructure:"output"`
	FilePath    string `mapstructure:"file_path"`
	MaxSizeMB   int    `mapstructure:"max_size_mb"`
	MaxBackups  int    `mapstructure:"max_backups"`
	MaxAgeDays  int    `mapstructure:"max_age_days"`
}

type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
	Path    string `mapstructure:"path"`
}

type HealthConfig struct {
	Enabled       bool   `mapstructure:"enabled"`
	Path          string `mapstructure:"path"`
	ReadinessPath string `mapstructure:"readiness_path"`
}

type ServiceDiscoveryConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	ConsulAddr  string `mapstructure:"consul_addr"`
	ServiceName string `mapstructure:"service_name"`
}

type SubscriptionsConfig struct {
	NATS        NATSConfig        `mapstructure:"nats"`
	RedisPubSub RedisPubSubConfig `mapstructure:"redis_pubsub"`
}

type NATSConfig struct {
	Enabled  bool     `mapstructure:"enabled"`
	URL      string   `mapstructure:"url"`
	Subjects []string `mapstructure:"subjects"`
}

type RedisPubSubConfig struct {
	Enabled  bool     `mapstructure:"enabled"`
	Channels []string `mapstructure:"channels"`
}

type AlertRulesConfig struct {
	Critical AlertRuleSettings `mapstructure:"critical"`
	Warning  AlertRuleSettings `mapstructure:"warning"`
	Info     AlertRuleSettings `mapstructure:"info"`
}

type AlertRuleSettings struct {
	Channels   []string      `mapstructure:"channels"`
	RetryMax   int           `mapstructure:"retry_max"`
	RetryDelay time.Duration `mapstructure:"retry_delay"`
}

type ServicesConfig struct {
	AccountMonitor  ServiceEndpoint `mapstructure:"account_monitor"`
	RiskManager     ServiceEndpoint `mapstructure:"risk_manager"`
	OrderExecution  ServiceEndpoint `mapstructure:"order_execution"`
}

type ServiceEndpoint struct {
	URL string `mapstructure:"url"`
}

// Load loads configuration from file and environment
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
		v.AddConfigPath("/etc/notification-service")
	}

	// Set defaults
	setDefaults(v)

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		// Config file is optional
		fmt.Printf("Warning: Config file not found: %v\n", err)
	}

	// Override with environment variables
	v.SetEnvPrefix("NOTIFICATION")
	v.AutomaticEnv()

	// Unmarshal config
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 9097)
	v.SetDefault("server.mode", "development")
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "30s")
	v.SetDefault("server.shutdown_timeout", "10s")

	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", "5m")

	// Redis defaults
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.max_retries", 3)
	v.SetDefault("redis.pool_size", 10)

	// Queue defaults
	v.SetDefault("queue.concurrency", 20)
	v.SetDefault("queue.max_retry", 3)
	v.SetDefault("queue.retry_delay_base", "5s")
	v.SetDefault("queue.retention_days", 7)

	// Preferences defaults
	v.SetDefault("preferences.default_channels", []string{"email"})
	v.SetDefault("preferences.opt_in_required", false)
	v.SetDefault("preferences.respect_quiet_hours", true)

	// Rate limit defaults
	v.SetDefault("rate_limit.email_per_hour", 50)
	v.SetDefault("rate_limit.sms_per_hour", 20)
	v.SetDefault("rate_limit.push_per_hour", 100)
	v.SetDefault("rate_limit.global_per_second", 100)

	// Templates defaults
	v.SetDefault("templates.directory", "./internal/templates")
	v.SetDefault("templates.auto_reload", true)
	v.SetDefault("templates.cache_enabled", true)

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.output", "stdout")

	// Metrics defaults
	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.port", 9098)
	v.SetDefault("metrics.path", "/metrics")

	// Health defaults
	v.SetDefault("health.enabled", true)
	v.SetDefault("health.path", "/health")
	v.SetDefault("health.readiness_path", "/ready")
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate server config
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	// Validate database config
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.DBName == "" {
		return fmt.Errorf("database name is required")
	}

	// Validate email config if email is enabled
	if c.Email.Provider != "" {
		if c.Email.FromAddress == "" {
			return fmt.Errorf("email from_address is required")
		}
		if c.Email.Provider == "sendgrid" && c.Email.SendGrid.APIKey == "" {
			return fmt.Errorf("sendgrid api_key is required")
		}
		if c.Email.Provider == "smtp" && c.Email.SMTP.Host == "" {
			return fmt.Errorf("smtp host is required")
		}
	}

	// Validate SMS config if SMS is enabled
	if c.SMS.Provider == "twilio" {
		if c.SMS.Twilio.AccountSID == "" || c.SMS.Twilio.AuthToken == "" {
			return fmt.Errorf("twilio credentials are required")
		}
	}

	// Validate push config if push is enabled
	if c.Push.Provider == "fcm" {
		if c.Push.FCM.CredentialsFile == "" {
			return fmt.Errorf("fcm credentials_file is required")
		}
	}

	return nil
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// GetRedisAddr returns the Redis address
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetServerAddr returns the server address
func (c *ServerConfig) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
