package quota

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/yourorg/b25/services/media/internal/config"
	"github.com/yourorg/b25/services/media/internal/models"
)

// QuotaManager manages storage quotas for users and organizations
type QuotaManager struct {
	db     *sql.DB
	config config.QuotaConfig
}

// NewQuotaManager creates a new quota manager
func NewQuotaManager(db *sql.DB, cfg config.QuotaConfig) *QuotaManager {
	return &QuotaManager{
		db:     db,
		config: cfg,
	}
}

// CheckQuota checks if an entity has enough quota for a file
func (m *QuotaManager) CheckQuota(entityID, entityType string, fileSize int64) (bool, error) {
	if !m.config.Enabled {
		return true, nil
	}

	quota, err := m.GetQuota(entityID, entityType)
	if err != nil {
		// If quota doesn't exist, create it with default
		if err := m.InitializeQuota(entityID, entityType); err != nil {
			return false, fmt.Errorf("failed to initialize quota: %w", err)
		}
		quota, err = m.GetQuota(entityID, entityType)
		if err != nil {
			return false, err
		}
	}

	// Check if adding this file would exceed quota
	if quota.Used+fileSize > quota.Limit {
		return false, nil
	}

	return true, nil
}

// UpdateQuota updates the used quota for an entity
func (m *QuotaManager) UpdateQuota(entityID, entityType string, delta int64) error {
	if !m.config.Enabled {
		return nil
	}

	query := `
		UPDATE media_quota
		SET used = used + $1, updated_at = $2
		WHERE entity_id = $3 AND type = $4
	`

	result, err := m.db.Exec(query, delta, time.Now(), entityID, entityType)
	if err != nil {
		return fmt.Errorf("failed to update quota: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("quota not found for entity %s", entityID)
	}

	return nil
}

// GetQuota retrieves the quota for an entity
func (m *QuotaManager) GetQuota(entityID, entityType string) (*models.MediaQuota, error) {
	query := `
		SELECT id, entity_id, type, used, "limit", updated_at
		FROM media_quota
		WHERE entity_id = $1 AND type = $2
	`

	var quota models.MediaQuota
	err := m.db.QueryRow(query, entityID, entityType).Scan(
		&quota.ID, &quota.EntityID, &quota.Type,
		&quota.Used, &quota.Limit, &quota.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("quota not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get quota: %w", err)
	}

	return &quota, nil
}

// InitializeQuota creates a new quota record with default limits
func (m *QuotaManager) InitializeQuota(entityID, entityType string) error {
	var limit int64
	if entityType == "user" {
		limit = m.config.DefaultUserQuota
	} else if entityType == "org" {
		limit = m.config.DefaultOrgQuota
	} else {
		return fmt.Errorf("invalid entity type: %s", entityType)
	}

	query := `
		INSERT INTO media_quota (id, entity_id, type, used, "limit")
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (entity_id, type) DO NOTHING
	`

	_, err := m.db.Exec(query, uuid.New().String(), entityID, entityType, 0, limit)
	if err != nil {
		return fmt.Errorf("failed to initialize quota: %w", err)
	}

	return nil
}

// SetQuotaLimit updates the quota limit for an entity
func (m *QuotaManager) SetQuotaLimit(entityID, entityType string, limit int64) error {
	query := `
		UPDATE media_quota
		SET "limit" = $1, updated_at = $2
		WHERE entity_id = $3 AND type = $4
	`

	result, err := m.db.Exec(query, limit, time.Now(), entityID, entityType)
	if err != nil {
		return fmt.Errorf("failed to set quota limit: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("quota not found for entity %s", entityID)
	}

	return nil
}

// RecalculateQuota recalculates the used quota from actual media files
func (m *QuotaManager) RecalculateQuota(entityID, entityType string) error {
	// Calculate actual usage
	var actualUsage int64
	query := `
		SELECT COALESCE(SUM(size), 0)
		FROM media
		WHERE deleted_at IS NULL
	`

	args := []interface{}{}
	if entityType == "user" {
		query += " AND user_id = $1"
		args = append(args, entityID)
	} else if entityType == "org" {
		query += " AND org_id = $1"
		args = append(args, entityID)
	} else {
		return fmt.Errorf("invalid entity type: %s", entityType)
	}

	err := m.db.QueryRow(query, args...).Scan(&actualUsage)
	if err != nil {
		return fmt.Errorf("failed to calculate usage: %w", err)
	}

	// Update quota with actual usage
	updateQuery := `
		UPDATE media_quota
		SET used = $1, updated_at = $2
		WHERE entity_id = $3 AND type = $4
	`

	_, err = m.db.Exec(updateQuery, actualUsage, time.Now(), entityID, entityType)
	if err != nil {
		return fmt.Errorf("failed to update quota: %w", err)
	}

	return nil
}

// StartPeriodicCheck starts a periodic quota recalculation
func (m *QuotaManager) StartPeriodicCheck(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(m.config.CheckInterval) * time.Second)
	defer ticker.Stop()

	log.Info("Starting periodic quota check")

	for {
		select {
		case <-ctx.Done():
			log.Info("Stopping periodic quota check")
			return
		case <-ticker.C:
			m.performPeriodicCheck()
		}
	}
}

// performPeriodicCheck performs a quota check for all entities
func (m *QuotaManager) performPeriodicCheck() {
	log.Debug("Performing periodic quota check")

	// Get all quota records
	query := `SELECT entity_id, type FROM media_quota`
	rows, err := m.db.Query(query)
	if err != nil {
		log.Errorf("Failed to query quotas: %v", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var entityID, entityType string
		if err := rows.Scan(&entityID, &entityType); err != nil {
			log.Errorf("Failed to scan quota row: %v", err)
			continue
		}

		if err := m.RecalculateQuota(entityID, entityType); err != nil {
			log.Errorf("Failed to recalculate quota for %s (%s): %v", entityID, entityType, err)
			continue
		}

		count++
	}

	log.Debugf("Recalculated %d quotas", count)
}

// GetQuotaUsagePercent returns the quota usage percentage
func (m *QuotaManager) GetQuotaUsagePercent(entityID, entityType string) (float64, error) {
	quota, err := m.GetQuota(entityID, entityType)
	if err != nil {
		return 0, err
	}

	if quota.Limit == 0 {
		return 0, nil
	}

	return float64(quota.Used) / float64(quota.Limit) * 100, nil
}

// IsQuotaExceeded checks if quota is exceeded
func (m *QuotaManager) IsQuotaExceeded(entityID, entityType string) (bool, error) {
	quota, err := m.GetQuota(entityID, entityType)
	if err != nil {
		return false, err
	}

	return quota.Used >= quota.Limit, nil
}
