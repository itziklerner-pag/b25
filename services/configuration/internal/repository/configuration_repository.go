package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/b25/services/configuration/internal/domain"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// ConfigurationRepository handles database operations for configurations
type ConfigurationRepository struct {
	db *sql.DB
}

// NewConfigurationRepository creates a new configuration repository
func NewConfigurationRepository(db *sql.DB) *ConfigurationRepository {
	return &ConfigurationRepository{db: db}
}

// Create creates a new configuration
func (r *ConfigurationRepository) Create(ctx context.Context, config *domain.Configuration) error {
	query := `
		INSERT INTO configurations (id, key, type, value, format, description, version, is_active, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(ctx, query,
		config.ID,
		config.Key,
		config.Type,
		config.Value,
		config.Format,
		config.Description,
		config.Version,
		config.IsActive,
		config.CreatedBy,
		config.CreatedAt,
		config.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return domain.ErrDuplicateKey
		}
		return fmt.Errorf("failed to create configuration: %w", err)
	}

	return nil
}

// GetByID retrieves a configuration by ID
func (r *ConfigurationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Configuration, error) {
	query := `
		SELECT id, key, type, value, format, description, version, is_active, created_by, created_at, updated_at
		FROM configurations
		WHERE id = $1
	`

	config := &domain.Configuration{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&config.ID,
		&config.Key,
		&config.Type,
		&config.Value,
		&config.Format,
		&config.Description,
		&config.Version,
		&config.IsActive,
		&config.CreatedBy,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration: %w", err)
	}

	return config, nil
}

// GetByKey retrieves a configuration by key
func (r *ConfigurationRepository) GetByKey(ctx context.Context, key string) (*domain.Configuration, error) {
	query := `
		SELECT id, key, type, value, format, description, version, is_active, created_by, created_at, updated_at
		FROM configurations
		WHERE key = $1
	`

	config := &domain.Configuration{}
	err := r.db.QueryRowContext(ctx, query, key).Scan(
		&config.ID,
		&config.Key,
		&config.Type,
		&config.Value,
		&config.Format,
		&config.Description,
		&config.Version,
		&config.IsActive,
		&config.CreatedBy,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration: %w", err)
	}

	return config, nil
}

// List retrieves configurations based on filter criteria
func (r *ConfigurationRepository) List(ctx context.Context, filter domain.ConfigurationFilter) ([]*domain.Configuration, error) {
	query := `
		SELECT id, key, type, value, format, description, version, is_active, created_by, created_at, updated_at
		FROM configurations
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 0

	if filter.Type != nil {
		argCount++
		query += fmt.Sprintf(" AND type = $%d", argCount)
		args = append(args, *filter.Type)
	}

	if filter.IsActive != nil {
		argCount++
		query += fmt.Sprintf(" AND is_active = $%d", argCount)
		args = append(args, *filter.IsActive)
	}

	if len(filter.Keys) > 0 {
		argCount++
		query += fmt.Sprintf(" AND key = ANY($%d)", argCount)
		args = append(args, pq.Array(filter.Keys))
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list configurations: %w", err)
	}
	defer rows.Close()

	var configs []*domain.Configuration
	for rows.Next() {
		config := &domain.Configuration{}
		err := rows.Scan(
			&config.ID,
			&config.Key,
			&config.Type,
			&config.Value,
			&config.Format,
			&config.Description,
			&config.Version,
			&config.IsActive,
			&config.CreatedBy,
			&config.CreatedAt,
			&config.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan configuration: %w", err)
		}
		configs = append(configs, config)
	}

	return configs, nil
}

// Update updates an existing configuration
func (r *ConfigurationRepository) Update(ctx context.Context, config *domain.Configuration) error {
	query := `
		UPDATE configurations
		SET value = $1, format = $2, description = $3, version = $4, updated_at = $5
		WHERE id = $6
	`

	result, err := r.db.ExecContext(ctx, query,
		config.Value,
		config.Format,
		config.Description,
		config.Version,
		config.UpdatedAt,
		config.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update configuration: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// UpdateStatus updates the active status of a configuration
func (r *ConfigurationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, isActive bool) error {
	query := `
		UPDATE configurations
		SET is_active = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, isActive, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update configuration status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// Delete deletes a configuration
func (r *ConfigurationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM configurations WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete configuration: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// CreateVersion creates a new configuration version
func (r *ConfigurationRepository) CreateVersion(ctx context.Context, version *domain.ConfigurationVersion) error {
	query := `
		INSERT INTO configuration_versions (id, configuration_id, version, value, format, changed_by, change_reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		version.ID,
		version.ConfigurationID,
		version.Version,
		version.Value,
		version.Format,
		version.ChangedBy,
		version.ChangeReason,
		version.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create configuration version: %w", err)
	}

	return nil
}

// GetVersions retrieves all versions of a configuration
func (r *ConfigurationRepository) GetVersions(ctx context.Context, configID uuid.UUID) ([]*domain.ConfigurationVersion, error) {
	query := `
		SELECT id, configuration_id, version, value, format, changed_by, change_reason, created_at
		FROM configuration_versions
		WHERE configuration_id = $1
		ORDER BY version DESC
	`

	rows, err := r.db.QueryContext(ctx, query, configID)
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration versions: %w", err)
	}
	defer rows.Close()

	var versions []*domain.ConfigurationVersion
	for rows.Next() {
		version := &domain.ConfigurationVersion{}
		err := rows.Scan(
			&version.ID,
			&version.ConfigurationID,
			&version.Version,
			&version.Value,
			&version.Format,
			&version.ChangedBy,
			&version.ChangeReason,
			&version.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan configuration version: %w", err)
		}
		versions = append(versions, version)
	}

	return versions, nil
}

// GetVersion retrieves a specific version of a configuration
func (r *ConfigurationRepository) GetVersion(ctx context.Context, configID uuid.UUID, version int) (*domain.ConfigurationVersion, error) {
	query := `
		SELECT id, configuration_id, version, value, format, changed_by, change_reason, created_at
		FROM configuration_versions
		WHERE configuration_id = $1 AND version = $2
	`

	v := &domain.ConfigurationVersion{}
	err := r.db.QueryRowContext(ctx, query, configID, version).Scan(
		&v.ID,
		&v.ConfigurationID,
		&v.Version,
		&v.Value,
		&v.Format,
		&v.ChangedBy,
		&v.ChangeReason,
		&v.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrInvalidVersion
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration version: %w", err)
	}

	return v, nil
}

// CreateAuditLog creates a new audit log entry
func (r *ConfigurationRepository) CreateAuditLog(ctx context.Context, log *domain.AuditLog) error {
	query := `
		INSERT INTO audit_logs (id, configuration_id, action, actor_id, actor_name, old_value, new_value, ip_address, user_agent, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(ctx, query,
		log.ID,
		log.ConfigurationID,
		log.Action,
		log.ActorID,
		log.ActorName,
		log.OldValue,
		log.NewValue,
		log.IPAddress,
		log.UserAgent,
		log.Timestamp,
	)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// GetAuditLogs retrieves audit logs for a configuration
func (r *ConfigurationRepository) GetAuditLogs(ctx context.Context, configID uuid.UUID, limit int) ([]*domain.AuditLog, error) {
	query := `
		SELECT id, configuration_id, action, actor_id, actor_name, old_value, new_value, ip_address, user_agent, timestamp
		FROM audit_logs
		WHERE configuration_id = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, configID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		log := &domain.AuditLog{}
		var oldValue, newValue sql.NullString

		err := rows.Scan(
			&log.ID,
			&log.ConfigurationID,
			&log.Action,
			&log.ActorID,
			&log.ActorName,
			&oldValue,
			&newValue,
			&log.IPAddress,
			&log.UserAgent,
			&log.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		if oldValue.Valid {
			log.OldValue = json.RawMessage(oldValue.String)
		}
		if newValue.Valid {
			log.NewValue = json.RawMessage(newValue.String)
		}

		logs = append(logs, log)
	}

	return logs, nil
}
