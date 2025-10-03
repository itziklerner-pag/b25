package domain

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ConfigType represents the type of configuration
type ConfigType string

const (
	ConfigTypeStrategy    ConfigType = "strategy"
	ConfigTypeRiskLimit   ConfigType = "risk_limit"
	ConfigTypeTradingPair ConfigType = "trading_pair"
	ConfigTypeSystem      ConfigType = "system"
)

// ConfigFormat represents the format of configuration data
type ConfigFormat string

const (
	ConfigFormatJSON ConfigFormat = "json"
	ConfigFormatYAML ConfigFormat = "yaml"
)

// Configuration represents a configuration entry
type Configuration struct {
	ID          uuid.UUID      `json:"id"`
	Key         string         `json:"key"`
	Type        ConfigType     `json:"type"`
	Value       json.RawMessage `json:"value"`
	Format      ConfigFormat   `json:"format"`
	Description string         `json:"description"`
	Version     int            `json:"version"`
	IsActive    bool           `json:"is_active"`
	CreatedBy   string         `json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// ConfigurationVersion represents a historical version of a configuration
type ConfigurationVersion struct {
	ID              uuid.UUID       `json:"id"`
	ConfigurationID uuid.UUID       `json:"configuration_id"`
	Version         int             `json:"version"`
	Value           json.RawMessage `json:"value"`
	Format          ConfigFormat    `json:"format"`
	ChangedBy       string          `json:"changed_by"`
	ChangeReason    string          `json:"change_reason"`
	CreatedAt       time.Time       `json:"created_at"`
}

// AuditLog represents an audit entry for configuration changes
type AuditLog struct {
	ID              uuid.UUID       `json:"id"`
	ConfigurationID uuid.UUID       `json:"configuration_id"`
	Action          string          `json:"action"`
	ActorID         string          `json:"actor_id"`
	ActorName       string          `json:"actor_name"`
	OldValue        json.RawMessage `json:"old_value,omitempty"`
	NewValue        json.RawMessage `json:"new_value,omitempty"`
	IPAddress       string          `json:"ip_address"`
	UserAgent       string          `json:"user_agent"`
	Timestamp       time.Time       `json:"timestamp"`
}

// CreateConfigurationRequest represents a request to create a configuration
type CreateConfigurationRequest struct {
	Key         string          `json:"key" binding:"required"`
	Type        ConfigType      `json:"type" binding:"required"`
	Value       json.RawMessage `json:"value" binding:"required"`
	Format      ConfigFormat    `json:"format" binding:"required"`
	Description string          `json:"description"`
	CreatedBy   string          `json:"created_by" binding:"required"`
}

// UpdateConfigurationRequest represents a request to update a configuration
type UpdateConfigurationRequest struct {
	Value        json.RawMessage `json:"value" binding:"required"`
	Format       ConfigFormat    `json:"format" binding:"required"`
	Description  string          `json:"description"`
	UpdatedBy    string          `json:"updated_by" binding:"required"`
	ChangeReason string          `json:"change_reason"`
}

// RollbackRequest represents a request to rollback to a specific version
type RollbackRequest struct {
	Version      int    `json:"version" binding:"required"`
	RolledBackBy string `json:"rolled_back_by" binding:"required"`
	Reason       string `json:"reason"`
}

// ConfigurationFilter represents filter criteria for listing configurations
type ConfigurationFilter struct {
	Type     *ConfigType
	IsActive *bool
	Keys     []string
	Limit    int
	Offset   int
}

// Validate validates the configuration type
func (ct ConfigType) Validate() error {
	switch ct {
	case ConfigTypeStrategy, ConfigTypeRiskLimit, ConfigTypeTradingPair, ConfigTypeSystem:
		return nil
	default:
		return fmt.Errorf("invalid config type: %s", ct)
	}
}

// Validate validates the configuration format
func (cf ConfigFormat) Validate() error {
	switch cf {
	case ConfigFormatJSON, ConfigFormatYAML:
		return nil
	default:
		return fmt.Errorf("invalid config format: %s", cf)
	}
}

// Validate validates the create configuration request
func (r *CreateConfigurationRequest) Validate() error {
	if r.Key == "" {
		return fmt.Errorf("key is required")
	}
	if err := r.Type.Validate(); err != nil {
		return err
	}
	if err := r.Format.Validate(); err != nil {
		return err
	}
	if len(r.Value) == 0 {
		return fmt.Errorf("value is required")
	}
	if r.CreatedBy == "" {
		return fmt.Errorf("created_by is required")
	}
	return nil
}

// Validate validates the update configuration request
func (r *UpdateConfigurationRequest) Validate() error {
	if err := r.Format.Validate(); err != nil {
		return err
	}
	if len(r.Value) == 0 {
		return fmt.Errorf("value is required")
	}
	if r.UpdatedBy == "" {
		return fmt.Errorf("updated_by is required")
	}
	return nil
}

// ConfigUpdateEvent represents a configuration update event for NATS
type ConfigUpdateEvent struct {
	ID        uuid.UUID       `json:"id"`
	Key       string          `json:"key"`
	Type      ConfigType      `json:"type"`
	Value     json.RawMessage `json:"value"`
	Format    ConfigFormat    `json:"format"`
	Version   int             `json:"version"`
	Action    string          `json:"action"` // created, updated, activated, deactivated, deleted
	Timestamp time.Time       `json:"timestamp"`
}
