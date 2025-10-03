package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/b25/services/configuration/internal/domain"
	"github.com/b25/services/configuration/internal/repository"
	"github.com/b25/services/configuration/internal/validator"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// ConfigurationService handles business logic for configurations
type ConfigurationService struct {
	repo      *repository.ConfigurationRepository
	validator *validator.Validator
	natsConn  *nats.Conn
	topic     string
	logger    *zap.Logger
}

// NewConfigurationService creates a new configuration service
func NewConfigurationService(
	repo *repository.ConfigurationRepository,
	validator *validator.Validator,
	natsConn *nats.Conn,
	topic string,
	logger *zap.Logger,
) *ConfigurationService {
	return &ConfigurationService{
		repo:      repo,
		validator: validator,
		natsConn:  natsConn,
		topic:     topic,
		logger:    logger,
	}
}

// Create creates a new configuration
func (s *ConfigurationService) Create(ctx context.Context, req *domain.CreateConfigurationRequest, actorName, ipAddress, userAgent string) (*domain.Configuration, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidInput, err)
	}

	// Validate configuration value
	if err := s.validator.Validate(req.Type, req.Format, req.Value); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrValidationFailed, err)
	}

	// Create configuration
	config := &domain.Configuration{
		ID:          uuid.New(),
		Key:         req.Key,
		Type:        req.Type,
		Value:       req.Value,
		Format:      req.Format,
		Description: req.Description,
		Version:     1,
		IsActive:    true,
		CreatedBy:   req.CreatedBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, config); err != nil {
		return nil, err
	}

	// Create initial version
	version := &domain.ConfigurationVersion{
		ID:              uuid.New(),
		ConfigurationID: config.ID,
		Version:         1,
		Value:           config.Value,
		Format:          config.Format,
		ChangedBy:       req.CreatedBy,
		ChangeReason:    "Initial creation",
		CreatedAt:       time.Now(),
	}

	if err := s.repo.CreateVersion(ctx, version); err != nil {
		s.logger.Error("Failed to create initial version", zap.Error(err))
	}

	// Create audit log
	auditLog := &domain.AuditLog{
		ID:              uuid.New(),
		ConfigurationID: config.ID,
		Action:          "created",
		ActorID:         req.CreatedBy,
		ActorName:       actorName,
		NewValue:        config.Value,
		IPAddress:       ipAddress,
		UserAgent:       userAgent,
		Timestamp:       time.Now(),
	}

	if err := s.repo.CreateAuditLog(ctx, auditLog); err != nil {
		s.logger.Error("Failed to create audit log", zap.Error(err))
	}

	// Publish update event
	if err := s.publishUpdateEvent(config, "created"); err != nil {
		s.logger.Error("Failed to publish update event", zap.Error(err))
	}

	return config, nil
}

// GetByID retrieves a configuration by ID
func (s *ConfigurationService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Configuration, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByKey retrieves a configuration by key
func (s *ConfigurationService) GetByKey(ctx context.Context, key string) (*domain.Configuration, error) {
	return s.repo.GetByKey(ctx, key)
}

// List retrieves configurations based on filter criteria
func (s *ConfigurationService) List(ctx context.Context, filter domain.ConfigurationFilter) ([]*domain.Configuration, error) {
	return s.repo.List(ctx, filter)
}

// Update updates an existing configuration
func (s *ConfigurationService) Update(ctx context.Context, id uuid.UUID, req *domain.UpdateConfigurationRequest, actorName, ipAddress, userAgent string) (*domain.Configuration, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidInput, err)
	}

	// Get existing configuration
	config, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate new value
	if err := s.validator.Validate(config.Type, req.Format, req.Value); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrValidationFailed, err)
	}

	// Store old value for audit
	oldValue := config.Value

	// Update configuration
	config.Value = req.Value
	config.Format = req.Format
	config.Description = req.Description
	config.Version++
	config.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, config); err != nil {
		return nil, err
	}

	// Create version
	version := &domain.ConfigurationVersion{
		ID:              uuid.New(),
		ConfigurationID: config.ID,
		Version:         config.Version,
		Value:           config.Value,
		Format:          config.Format,
		ChangedBy:       req.UpdatedBy,
		ChangeReason:    req.ChangeReason,
		CreatedAt:       time.Now(),
	}

	if err := s.repo.CreateVersion(ctx, version); err != nil {
		s.logger.Error("Failed to create version", zap.Error(err))
	}

	// Create audit log
	auditLog := &domain.AuditLog{
		ID:              uuid.New(),
		ConfigurationID: config.ID,
		Action:          "updated",
		ActorID:         req.UpdatedBy,
		ActorName:       actorName,
		OldValue:        oldValue,
		NewValue:        config.Value,
		IPAddress:       ipAddress,
		UserAgent:       userAgent,
		Timestamp:       time.Now(),
	}

	if err := s.repo.CreateAuditLog(ctx, auditLog); err != nil {
		s.logger.Error("Failed to create audit log", zap.Error(err))
	}

	// Publish update event
	if err := s.publishUpdateEvent(config, "updated"); err != nil {
		s.logger.Error("Failed to publish update event", zap.Error(err))
	}

	return config, nil
}

// Activate activates a configuration
func (s *ConfigurationService) Activate(ctx context.Context, id uuid.UUID, actorID, actorName, ipAddress, userAgent string) error {
	config, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.UpdateStatus(ctx, id, true); err != nil {
		return err
	}

	config.IsActive = true

	// Create audit log
	auditLog := &domain.AuditLog{
		ID:              uuid.New(),
		ConfigurationID: id,
		Action:          "activated",
		ActorID:         actorID,
		ActorName:       actorName,
		IPAddress:       ipAddress,
		UserAgent:       userAgent,
		Timestamp:       time.Now(),
	}

	if err := s.repo.CreateAuditLog(ctx, auditLog); err != nil {
		s.logger.Error("Failed to create audit log", zap.Error(err))
	}

	// Publish update event
	if err := s.publishUpdateEvent(config, "activated"); err != nil {
		s.logger.Error("Failed to publish update event", zap.Error(err))
	}

	return nil
}

// Deactivate deactivates a configuration
func (s *ConfigurationService) Deactivate(ctx context.Context, id uuid.UUID, actorID, actorName, ipAddress, userAgent string) error {
	config, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.UpdateStatus(ctx, id, false); err != nil {
		return err
	}

	config.IsActive = false

	// Create audit log
	auditLog := &domain.AuditLog{
		ID:              uuid.New(),
		ConfigurationID: id,
		Action:          "deactivated",
		ActorID:         actorID,
		ActorName:       actorName,
		IPAddress:       ipAddress,
		UserAgent:       userAgent,
		Timestamp:       time.Now(),
	}

	if err := s.repo.CreateAuditLog(ctx, auditLog); err != nil {
		s.logger.Error("Failed to create audit log", zap.Error(err))
	}

	// Publish update event
	if err := s.publishUpdateEvent(config, "deactivated"); err != nil {
		s.logger.Error("Failed to publish update event", zap.Error(err))
	}

	return nil
}

// Delete deletes a configuration
func (s *ConfigurationService) Delete(ctx context.Context, id uuid.UUID, actorID, actorName, ipAddress, userAgent string) error {
	config, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Create audit log
	auditLog := &domain.AuditLog{
		ID:              uuid.New(),
		ConfigurationID: id,
		Action:          "deleted",
		ActorID:         actorID,
		ActorName:       actorName,
		OldValue:        config.Value,
		IPAddress:       ipAddress,
		UserAgent:       userAgent,
		Timestamp:       time.Now(),
	}

	if err := s.repo.CreateAuditLog(ctx, auditLog); err != nil {
		s.logger.Error("Failed to create audit log", zap.Error(err))
	}

	// Publish update event
	if err := s.publishUpdateEvent(config, "deleted"); err != nil {
		s.logger.Error("Failed to publish update event", zap.Error(err))
	}

	return nil
}

// GetVersions retrieves all versions of a configuration
func (s *ConfigurationService) GetVersions(ctx context.Context, id uuid.UUID) ([]*domain.ConfigurationVersion, error) {
	return s.repo.GetVersions(ctx, id)
}

// Rollback rolls back a configuration to a specific version
func (s *ConfigurationService) Rollback(ctx context.Context, id uuid.UUID, req *domain.RollbackRequest, actorName, ipAddress, userAgent string) (*domain.Configuration, error) {
	// Get current configuration
	config, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get target version
	targetVersion, err := s.repo.GetVersion(ctx, id, req.Version)
	if err != nil {
		return nil, err
	}

	// Store old value for audit
	oldValue := config.Value

	// Update configuration with version data
	config.Value = targetVersion.Value
	config.Format = targetVersion.Format
	config.Version++
	config.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, config); err != nil {
		return nil, err
	}

	// Create new version entry for rollback
	version := &domain.ConfigurationVersion{
		ID:              uuid.New(),
		ConfigurationID: config.ID,
		Version:         config.Version,
		Value:           config.Value,
		Format:          config.Format,
		ChangedBy:       req.RolledBackBy,
		ChangeReason:    fmt.Sprintf("Rollback to version %d: %s", req.Version, req.Reason),
		CreatedAt:       time.Now(),
	}

	if err := s.repo.CreateVersion(ctx, version); err != nil {
		s.logger.Error("Failed to create version", zap.Error(err))
	}

	// Create audit log
	auditLog := &domain.AuditLog{
		ID:              uuid.New(),
		ConfigurationID: config.ID,
		Action:          fmt.Sprintf("rolled_back_to_v%d", req.Version),
		ActorID:         req.RolledBackBy,
		ActorName:       actorName,
		OldValue:        oldValue,
		NewValue:        config.Value,
		IPAddress:       ipAddress,
		UserAgent:       userAgent,
		Timestamp:       time.Now(),
	}

	if err := s.repo.CreateAuditLog(ctx, auditLog); err != nil {
		s.logger.Error("Failed to create audit log", zap.Error(err))
	}

	// Publish update event
	if err := s.publishUpdateEvent(config, "updated"); err != nil {
		s.logger.Error("Failed to publish update event", zap.Error(err))
	}

	return config, nil
}

// GetAuditLogs retrieves audit logs for a configuration
func (s *ConfigurationService) GetAuditLogs(ctx context.Context, id uuid.UUID, limit int) ([]*domain.AuditLog, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.repo.GetAuditLogs(ctx, id, limit)
}

// publishUpdateEvent publishes a configuration update event to NATS
func (s *ConfigurationService) publishUpdateEvent(config *domain.Configuration, action string) error {
	if s.natsConn == nil {
		return fmt.Errorf("NATS connection not available")
	}

	event := &domain.ConfigUpdateEvent{
		ID:        config.ID,
		Key:       config.Key,
		Type:      config.Type,
		Value:     config.Value,
		Format:    config.Format,
		Version:   config.Version,
		Action:    action,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	subject := fmt.Sprintf("%s.%s", s.topic, config.Type)
	if err := s.natsConn.Publish(subject, data); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	s.logger.Info("Published config update event",
		zap.String("subject", subject),
		zap.String("key", config.Key),
		zap.String("action", action),
	)

	return nil
}
