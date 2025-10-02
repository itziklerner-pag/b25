package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "Valid configuration",
			config: Config{
				Server: ServerConfig{
					Port: 9097,
				},
				Database: DatabaseConfig{
					Host:     "localhost",
					Database: "analytics",
				},
				Kafka: KafkaConfig{
					Brokers: []string{"localhost:9092"},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid port",
			config: Config{
				Server: ServerConfig{
					Port: 99999,
				},
			},
			wantErr: true,
		},
		{
			name: "Missing database host",
			config: Config{
				Server: ServerConfig{
					Port: 9097,
				},
				Database: DatabaseConfig{
					Database: "analytics",
				},
			},
			wantErr: true,
		},
		{
			name: "Missing Kafka brokers",
			config: Config{
				Server: ServerConfig{
					Port: 9097,
				},
				Database: DatabaseConfig{
					Host:     "localhost",
					Database: "analytics",
				},
				Kafka: KafkaConfig{
					Brokers: []string{},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigLoading(t *testing.T) {
	// Create a temporary config file
	configData := `
server:
  host: "0.0.0.0"
  port: 9097
  read_timeout: 30s
  write_timeout: 30s
  shutdown_timeout: 10s

database:
  host: "localhost"
  port: 5432
  database: "test_analytics"
  user: "test_user"
  password: "test_password"
  ssl_mode: "disable"
  max_connections: 10
  max_idle_connections: 5
  connection_lifetime: 300s

redis:
  host: "localhost"
  port: 6379
  db: 0
  pool_size: 10

kafka:
  brokers:
    - "localhost:9092"
  consumer_group: "test-group"
  topics:
    - "test.events"

analytics:
  ingestion:
    batch_size: 100
    batch_timeout: 5s
    workers: 2
    buffer_size: 1000
  aggregation:
    intervals:
      - "1m"
      - "1h"
    workers: 1
  retention:
    raw_events: 90d
  query:
    max_results: 1000
    default_limit: 100
    timeout: 30s
    cache_ttl: 60s

metrics:
  enabled: true
  port: 9098

logging:
  level: "info"
  format: "json"
`

	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configData)
	assert.NoError(t, err)
	tmpFile.Close()

	// Load config
	cfg, err := Load(tmpFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// Verify values
	assert.Equal(t, 9097, cfg.Server.Port)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, "test_analytics", cfg.Database.Database)
	assert.Equal(t, 1, len(cfg.Kafka.Brokers))
	assert.Equal(t, 100, cfg.Analytics.Ingestion.BatchSize)
	assert.Equal(t, 2, len(cfg.Analytics.Aggregation.Intervals))
}

func TestEnvironmentOverrides(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port: 9097,
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "analytics",
		},
	}

	// Set environment variables
	os.Setenv("SERVER_PORT", "8080")
	os.Setenv("DB_HOST", "postgres.example.com")
	os.Setenv("DB_PORT", "5433")
	defer os.Unsetenv("SERVER_PORT")
	defer os.Unsetenv("DB_HOST")
	defer os.Unsetenv("DB_PORT")

	cfg.overrideFromEnv()

	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "postgres.example.com", cfg.Database.Host)
	assert.Equal(t, 5433, cfg.Database.Port)
}

func TestYAMLMarshaling(t *testing.T) {
	cfg := Config{
		Server: ServerConfig{
			Host:            "0.0.0.0",
			Port:            9097,
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			ShutdownTimeout: 10 * time.Second,
		},
	}

	data, err := yaml.Marshal(&cfg)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	var decoded Config
	err = yaml.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, cfg.Server.Port, decoded.Server.Port)
}
