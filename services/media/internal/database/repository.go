package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/yourorg/b25/services/media/internal/models"
)

// MediaRepository handles database operations for media
type MediaRepository struct {
	db *sql.DB
}

// NewMediaRepository creates a new media repository
func NewMediaRepository(db *sql.DB) *MediaRepository {
	return &MediaRepository{db: db}
}

// Create inserts a new media record
func (r *MediaRepository) Create(media *models.Media) error {
	metadataJSON, err := json.Marshal(media.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO media (
			id, user_id, org_id, file_name, original_name, mime_type,
			media_type, size, status, storage_path, public_url, cdn_url, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err = r.db.Exec(
		query,
		media.ID, media.UserID, media.OrgID, media.FileName, media.OriginalName,
		media.MimeType, media.MediaType, media.Size, media.Status, media.StoragePath,
		media.PublicURL, media.CDNUrl, metadataJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to create media: %w", err)
	}

	return nil
}

// GetByID retrieves a media record by ID
func (r *MediaRepository) GetByID(id string) (*models.Media, error) {
	query := `
		SELECT id, user_id, org_id, file_name, original_name, mime_type,
			   media_type, size, status, storage_path, public_url, cdn_url,
			   metadata, created_at, updated_at, deleted_at
		FROM media
		WHERE id = $1 AND deleted_at IS NULL
	`

	var media models.Media
	var metadataJSON []byte
	var orgID, cdnURL sql.NullString
	var deletedAt sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&media.ID, &media.UserID, &orgID, &media.FileName, &media.OriginalName,
		&media.MimeType, &media.MediaType, &media.Size, &media.Status,
		&media.StoragePath, &media.PublicURL, &cdnURL, &metadataJSON,
		&media.CreatedAt, &media.UpdatedAt, &deletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("media not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get media: %w", err)
	}

	if orgID.Valid {
		media.OrgID = orgID.String
	}
	if cdnURL.Valid {
		media.CDNUrl = cdnURL.String
	}
	if deletedAt.Valid {
		media.DeletedAt = &deletedAt.Time
	}

	if err := json.Unmarshal(metadataJSON, &media.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &media, nil
}

// List retrieves media records with filters
func (r *MediaRepository) List(filter models.MediaFilter) ([]*models.Media, error) {
	query := `
		SELECT id, user_id, org_id, file_name, original_name, mime_type,
			   media_type, size, status, storage_path, public_url, cdn_url,
			   metadata, created_at, updated_at
		FROM media
		WHERE deleted_at IS NULL
	`

	args := []interface{}{}
	argCount := 1

	if filter.UserID != "" {
		query += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, filter.UserID)
		argCount++
	}

	if filter.OrgID != "" {
		query += fmt.Sprintf(" AND org_id = $%d", argCount)
		args = append(args, filter.OrgID)
		argCount++
	}

	if filter.MediaType != "" {
		query += fmt.Sprintf(" AND media_type = $%d", argCount)
		args = append(args, filter.MediaType)
		argCount++
	}

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, filter.Status)
		argCount++
	}

	// Sorting
	if filter.SortBy != "" {
		order := "ASC"
		if filter.SortOrder == "desc" {
			order = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", filter.SortBy, order)
	} else {
		query += " ORDER BY created_at DESC"
	}

	// Pagination
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)
		argCount++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filter.Offset)
		argCount++
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list media: %w", err)
	}
	defer rows.Close()

	var mediaList []*models.Media

	for rows.Next() {
		var media models.Media
		var metadataJSON []byte
		var orgID, cdnURL sql.NullString

		err := rows.Scan(
			&media.ID, &media.UserID, &orgID, &media.FileName, &media.OriginalName,
			&media.MimeType, &media.MediaType, &media.Size, &media.Status,
			&media.StoragePath, &media.PublicURL, &cdnURL, &metadataJSON,
			&media.CreatedAt, &media.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan media: %w", err)
		}

		if orgID.Valid {
			media.OrgID = orgID.String
		}
		if cdnURL.Valid {
			media.CDNUrl = cdnURL.String
		}

		if err := json.Unmarshal(metadataJSON, &media.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		mediaList = append(mediaList, &media)
	}

	return mediaList, nil
}

// UpdateStatus updates the status of a media record
func (r *MediaRepository) UpdateStatus(id string, status models.MediaStatus) error {
	query := `UPDATE media SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.Exec(query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	return nil
}

// UpdateMetadata updates the metadata of a media record
func (r *MediaRepository) UpdateMetadata(id string, metadata models.Metadata) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `UPDATE media SET metadata = $1, updated_at = $2 WHERE id = $3`
	_, err = r.db.Exec(query, metadataJSON, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}
	return nil
}

// Delete soft deletes a media record
func (r *MediaRepository) Delete(id string) error {
	query := `UPDATE media SET deleted_at = $1 WHERE id = $2`
	_, err := r.db.Exec(query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete media: %w", err)
	}
	return nil
}

// GetStats retrieves statistics for a user or organization
func (r *MediaRepository) GetStats(entityID string, entityType string) (*models.MediaStats, error) {
	stats := &models.MediaStats{
		ByType:   make(map[models.MediaType]int),
		ByStatus: make(map[models.MediaStatus]int),
	}

	// Total files and size
	query := `
		SELECT COUNT(*), COALESCE(SUM(size), 0)
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
	}

	err := r.db.QueryRow(query, args...).Scan(&stats.TotalFiles, &stats.TotalSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	// By type
	typeQuery := query + " GROUP BY media_type"
	// By status
	statusQuery := query + " GROUP BY status"

	// Get quota information
	quotaQuery := `SELECT used, "limit" FROM media_quota WHERE entity_id = $1 AND type = $2`
	err = r.db.QueryRow(quotaQuery, entityID, entityType).Scan(&stats.QuotaUsed, &stats.QuotaLimit)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get quota: %w", err)
	}

	if stats.QuotaLimit > 0 {
		stats.QuotaPercent = float64(stats.QuotaUsed) / float64(stats.QuotaLimit) * 100
	}

	return stats, nil
}
