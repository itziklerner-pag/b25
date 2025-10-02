package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config holds all configuration for the media service
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Storage    StorageConfig    `mapstructure:"storage"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	CDN        CDNConfig        `mapstructure:"cdn"`
	Processing ProcessingConfig `mapstructure:"processing"`
	Security   SecurityConfig   `mapstructure:"security"`
	Quota      QuotaConfig      `mapstructure:"quota"`
	Streaming  StreamingConfig  `mapstructure:"streaming"`
	Metrics    MetricsConfig    `mapstructure:"metrics"`
	Health     HealthConfig     `mapstructure:"health"`
	Logging    LoggingConfig    `mapstructure:"logging"`
}

type ServerConfig struct {
	Port          int    `mapstructure:"port"`
	Host          string `mapstructure:"host"`
	Mode          string `mapstructure:"mode"`
	UploadPath    string `mapstructure:"upload_path"`
	MaxUploadSize int64  `mapstructure:"max_upload_size"`
}

type StorageConfig struct {
	Type  string       `mapstructure:"type"`
	S3    S3Config     `mapstructure:"s3"`
	Local LocalConfig  `mapstructure:"local"`
}

type S3Config struct {
	Endpoint       string `mapstructure:"endpoint"`
	Region         string `mapstructure:"region"`
	Bucket         string `mapstructure:"bucket"`
	AccessKey      string `mapstructure:"access_key"`
	SecretKey      string `mapstructure:"secret_key"`
	UseSSL         bool   `mapstructure:"use_ssl"`
	ForcePathStyle bool   `mapstructure:"force_path_style"`
}

type LocalConfig struct {
	BasePath string `mapstructure:"base_path"`
}

type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	SSLMode         string `mapstructure:"sslmode"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type CDNConfig struct {
	Enabled       bool   `mapstructure:"enabled"`
	BaseURL       string `mapstructure:"base_url"`
	CacheTTL      int    `mapstructure:"cache_ttl"`
	PurgeOnDelete bool   `mapstructure:"purge_on_delete"`
}

type ProcessingConfig struct {
	Image ImageConfig `mapstructure:"image"`
	Video VideoConfig `mapstructure:"video"`
}

type ImageConfig struct {
	MaxWidth   int               `mapstructure:"max_width"`
	MaxHeight  int               `mapstructure:"max_height"`
	Quality    int               `mapstructure:"quality"`
	Formats    []string          `mapstructure:"formats"`
	Thumbnails []ThumbnailConfig `mapstructure:"thumbnails"`
}

type ThumbnailConfig struct {
	Name   string `mapstructure:"name"`
	Width  int    `mapstructure:"width"`
	Height int    `mapstructure:"height"`
}

type VideoConfig struct {
	MaxDuration   int              `mapstructure:"max_duration"`
	MaxSize       int64            `mapstructure:"max_size"`
	OutputFormats []string         `mapstructure:"output_formats"`
	Codecs        CodecsConfig     `mapstructure:"codecs"`
	Profiles      []ProfileConfig  `mapstructure:"profiles"`
	Thumbnail     VideoThumbConfig `mapstructure:"thumbnail"`
}

type CodecsConfig struct {
	Video string `mapstructure:"video"`
	Audio string `mapstructure:"audio"`
}

type ProfileConfig struct {
	Name    string `mapstructure:"name"`
	Width   int    `mapstructure:"width"`
	Height  int    `mapstructure:"height"`
	Bitrate string `mapstructure:"bitrate"`
}

type VideoThumbConfig struct {
	Time   string `mapstructure:"time"`
	Width  int    `mapstructure:"width"`
	Height int    `mapstructure:"height"`
}

type SecurityConfig struct {
	EnableVirusScan  bool          `mapstructure:"enable_virus_scan"`
	ClamAV           ClamAVConfig  `mapstructure:"clamav"`
	AllowedMimeTypes []string      `mapstructure:"allowed_mime_types"`
	MaxFileSize      int64         `mapstructure:"max_file_size"`
}

type ClamAVConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type QuotaConfig struct {
	Enabled          bool  `mapstructure:"enabled"`
	DefaultUserQuota int64 `mapstructure:"default_user_quota"`
	DefaultOrgQuota  int64 `mapstructure:"default_org_quota"`
	CheckInterval    int   `mapstructure:"check_interval"`
}

type StreamingConfig struct {
	Enabled            bool     `mapstructure:"enabled"`
	ChunkSize          int      `mapstructure:"chunk_size"`
	SupportedProtocols []string `mapstructure:"supported_protocols"`
}

type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
}

type HealthConfig struct {
	Path string `mapstructure:"path"`
}

type LoggingConfig struct {
	Level    string `mapstructure:"level"`
	Format   string `mapstructure:"format"`
	Output   string `mapstructure:"output"`
	FilePath string `mapstructure:"file_path"`
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	v := viper.New()

	// Set config file locations
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("/etc/media-service")

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found; use defaults and env vars
	}

	// Environment variable overrides
	v.AutomaticEnv()
	v.SetEnvPrefix("MEDIA")

	// Map environment variables to config keys
	bindEnvVars(v)

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate configuration
	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

func bindEnvVars(v *viper.Viper) {
	// Server
	v.BindEnv("server.port", "SERVER_PORT")
	v.BindEnv("server.host", "SERVER_HOST")
	v.BindEnv("server.mode", "SERVER_MODE")

	// Storage
	v.BindEnv("storage.type", "STORAGE_TYPE")
	v.BindEnv("storage.s3.endpoint", "S3_ENDPOINT")
	v.BindEnv("storage.s3.region", "S3_REGION")
	v.BindEnv("storage.s3.bucket", "S3_BUCKET")
	v.BindEnv("storage.s3.access_key", "S3_ACCESS_KEY")
	v.BindEnv("storage.s3.secret_key", "S3_SECRET_KEY")
	v.BindEnv("storage.s3.use_ssl", "S3_USE_SSL")

	// Database
	v.BindEnv("database.host", "DB_HOST")
	v.BindEnv("database.port", "DB_PORT")
	v.BindEnv("database.user", "DB_USER")
	v.BindEnv("database.password", "DB_PASSWORD")
	v.BindEnv("database.dbname", "DB_NAME")
	v.BindEnv("database.sslmode", "DB_SSLMODE")

	// Redis
	v.BindEnv("redis.host", "REDIS_HOST")
	v.BindEnv("redis.port", "REDIS_PORT")
	v.BindEnv("redis.password", "REDIS_PASSWORD")
	v.BindEnv("redis.db", "REDIS_DB")

	// Security
	v.BindEnv("security.enable_virus_scan", "ENABLE_VIRUS_SCAN")
	v.BindEnv("security.clamav.host", "CLAMAV_HOST")
	v.BindEnv("security.clamav.port", "CLAMAV_PORT")

	// Quota
	v.BindEnv("quota.enabled", "QUOTA_ENABLED")
	v.BindEnv("quota.default_user_quota", "DEFAULT_USER_QUOTA")
	v.BindEnv("quota.default_org_quota", "DEFAULT_ORG_QUOTA")

	// Logging
	v.BindEnv("logging.level", "LOG_LEVEL")
	v.BindEnv("logging.format", "LOG_FORMAT")
}

func validate(cfg *Config) error {
	// Validate server config
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", cfg.Server.Port)
	}

	// Validate storage config
	if cfg.Storage.Type != "s3" && cfg.Storage.Type != "local" {
		return fmt.Errorf("invalid storage type: %s", cfg.Storage.Type)
	}

	if cfg.Storage.Type == "s3" {
		if cfg.Storage.S3.Bucket == "" {
			return fmt.Errorf("S3 bucket name is required")
		}
		if cfg.Storage.S3.Region == "" {
			return fmt.Errorf("S3 region is required")
		}
	}

	// Validate database config
	if cfg.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	// Create upload directory if it doesn't exist
	if cfg.Server.UploadPath != "" {
		if err := os.MkdirAll(cfg.Server.UploadPath, 0755); err != nil {
			return fmt.Errorf("failed to create upload directory: %w", err)
		}
	}

	return nil
}
