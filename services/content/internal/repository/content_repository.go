package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/b25/services/content/internal/domain"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type ContentRepository interface {
	Create(ctx context.Context, content *domain.Content) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Content, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Content, error)
	Update(ctx context.Context, id uuid.UUID, content *domain.Content) error
	Delete(ctx context.Context, id uuid.UUID) error
	Search(ctx context.Context, params domain.SearchContentParams) (*domain.PaginatedContent, error)
	IncrementViewCount(ctx context.Context, id uuid.UUID) error
	GetVersions(ctx context.Context, contentID uuid.UUID) ([]domain.ContentVersion, error)
	GetVersion(ctx context.Context, contentID uuid.UUID, version int) (*domain.ContentVersion, error)
}

type contentRepository struct {
	db *sql.DB
}

func NewContentRepository(db *sql.DB) ContentRepository {
	return &contentRepository{db: db}
}

func (r *contentRepository) Create(ctx context.Context, content *domain.Content) error {
	query := `
		INSERT INTO content (
			id, type, title, slug, body, excerpt, status, author_id,
			tags, categories, metadata, media_url, media_type, media_size,
			published_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING version, created_at, updated_at
	`

	metadataJSON, err := json.Marshal(content.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	err = r.db.QueryRowContext(
		ctx, query,
		content.ID, content.Type, content.Title, content.Slug,
		content.Body, content.Excerpt, content.Status, content.AuthorID,
		pq.Array(content.Tags), pq.Array(content.Categories), metadataJSON,
		content.MediaURL, content.MediaType, content.MediaSize,
		content.PublishedAt, time.Now(), time.Now(),
	).Scan(&content.Version, &content.CreatedAt, &content.UpdatedAt)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return domain.ErrContentExists
		}
		return fmt.Errorf("failed to create content: %w", err)
	}

	return nil
}

func (r *contentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Content, error) {
	query := `
		SELECT
			c.id, c.type, c.title, c.slug, c.body, c.excerpt, c.status,
			c.author_id, c.tags, c.categories, c.metadata, c.media_url,
			c.media_type, c.media_size, c.version, c.view_count,
			c.published_at, c.created_at, c.updated_at,
			u.username as author_name
		FROM content c
		JOIN users u ON c.author_id = u.id
		WHERE c.id = $1
	`

	content := &domain.Content{
		Tags:       []string{},
		Categories: []string{},
		Metadata:   make(map[string]interface{}),
	}

	var metadataJSON []byte
	var publishedAt sql.NullTime
	var mediaURL, mediaType sql.NullString
	var mediaSize sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&content.ID, &content.Type, &content.Title, &content.Slug,
		&content.Body, &content.Excerpt, &content.Status, &content.AuthorID,
		pq.Array(&content.Tags), pq.Array(&content.Categories), &metadataJSON,
		&mediaURL, &mediaType, &mediaSize, &content.Version, &content.ViewCount,
		&publishedAt, &content.CreatedAt, &content.UpdatedAt, &content.AuthorName,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrContentNotFound
		}
		return nil, fmt.Errorf("failed to get content: %w", err)
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &content.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	if publishedAt.Valid {
		content.PublishedAt = &publishedAt.Time
	}
	if mediaURL.Valid {
		content.MediaURL = mediaURL.String
	}
	if mediaType.Valid {
		content.MediaType = mediaType.String
	}
	if mediaSize.Valid {
		content.MediaSize = mediaSize.Int64
	}

	return content, nil
}

func (r *contentRepository) GetBySlug(ctx context.Context, slug string) (*domain.Content, error) {
	query := `
		SELECT
			c.id, c.type, c.title, c.slug, c.body, c.excerpt, c.status,
			c.author_id, c.tags, c.categories, c.metadata, c.media_url,
			c.media_type, c.media_size, c.version, c.view_count,
			c.published_at, c.created_at, c.updated_at,
			u.username as author_name
		FROM content c
		JOIN users u ON c.author_id = u.id
		WHERE c.slug = $1
	`

	content := &domain.Content{
		Tags:       []string{},
		Categories: []string{},
		Metadata:   make(map[string]interface{}),
	}

	var metadataJSON []byte
	var publishedAt sql.NullTime
	var mediaURL, mediaType sql.NullString
	var mediaSize sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&content.ID, &content.Type, &content.Title, &content.Slug,
		&content.Body, &content.Excerpt, &content.Status, &content.AuthorID,
		pq.Array(&content.Tags), pq.Array(&content.Categories), &metadataJSON,
		&mediaURL, &mediaType, &mediaSize, &content.Version, &content.ViewCount,
		&publishedAt, &content.CreatedAt, &content.UpdatedAt, &content.AuthorName,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrContentNotFound
		}
		return nil, fmt.Errorf("failed to get content by slug: %w", err)
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &content.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	if publishedAt.Valid {
		content.PublishedAt = &publishedAt.Time
	}
	if mediaURL.Valid {
		content.MediaURL = mediaURL.String
	}
	if mediaType.Valid {
		content.MediaType = mediaType.String
	}
	if mediaSize.Valid {
		content.MediaSize = mediaSize.Int64
	}

	return content, nil
}

func (r *contentRepository) Update(ctx context.Context, id uuid.UUID, content *domain.Content) error {
	query := `
		UPDATE content SET
			title = $2, slug = $3, body = $4, excerpt = $5, status = $6,
			tags = $7, categories = $8, metadata = $9,
			media_url = $10, media_type = $11, media_size = $12,
			published_at = $13, updated_at = NOW()
		WHERE id = $1
		RETURNING version, updated_at
	`

	metadataJSON, err := json.Marshal(content.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	err = r.db.QueryRowContext(
		ctx, query,
		id, content.Title, content.Slug, content.Body, content.Excerpt, content.Status,
		pq.Array(content.Tags), pq.Array(content.Categories), metadataJSON,
		content.MediaURL, content.MediaType, content.MediaSize, content.PublishedAt,
	).Scan(&content.Version, &content.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrContentNotFound
		}
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return domain.ErrContentExists
		}
		return fmt.Errorf("failed to update content: %w", err)
	}

	return nil
}

func (r *contentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM content WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete content: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return domain.ErrContentNotFound
	}

	return nil
}

func (r *contentRepository) Search(ctx context.Context, params domain.SearchContentParams) (*domain.PaginatedContent, error) {
	// Build WHERE clause
	var whereClauses []string
	var args []interface{}
	argIndex := 1

	// Full-text search
	if params.Query != "" {
		whereClauses = append(whereClauses, fmt.Sprintf(
			"to_tsvector('english', COALESCE(c.title, '') || ' ' || COALESCE(c.body, '') || ' ' || COALESCE(c.excerpt, '')) @@ plainto_tsquery('english', $%d)",
			argIndex,
		))
		args = append(args, params.Query)
		argIndex++
	}

	// Type filter
	if params.Type != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("c.type = $%d", argIndex))
		args = append(args, params.Type)
		argIndex++
	}

	// Status filter
	if params.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("c.status = $%d", argIndex))
		args = append(args, params.Status)
		argIndex++
	}

	// Author filter
	if params.AuthorID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("c.author_id = $%d", argIndex))
		args = append(args, *params.AuthorID)
		argIndex++
	}

	// Tags filter
	if len(params.Tags) > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("c.tags && $%d", argIndex))
		args = append(args, pq.Array(params.Tags))
		argIndex++
	}

	// Categories filter
	if len(params.Categories) > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("c.categories && $%d", argIndex))
		args = append(args, pq.Array(params.Categories))
		argIndex++
	}

	// Date range filters
	if params.FromDate != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("c.created_at >= $%d", argIndex))
		args = append(args, *params.FromDate)
		argIndex++
	}
	if params.ToDate != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("c.created_at <= $%d", argIndex))
		args = append(args, *params.ToDate)
		argIndex++
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Sorting
	sortBy := "c.created_at"
	if params.SortBy != "" {
		switch params.SortBy {
		case "updated_at", "view_count", "title":
			sortBy = "c." + params.SortBy
		case "created_at":
			sortBy = "c.created_at"
		}
	}

	sortOrder := "DESC"
	if params.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	// Pagination
	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	// Count total
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM content c
		%s
	`, whereClause)

	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count content: %w", err)
	}

	// Fetch items
	query := fmt.Sprintf(`
		SELECT
			c.id, c.type, c.title, c.slug, c.body, c.excerpt, c.status,
			c.author_id, c.tags, c.categories, c.metadata, c.media_url,
			c.media_type, c.media_size, c.version, c.view_count,
			c.published_at, c.created_at, c.updated_at,
			u.username as author_name
		FROM content c
		JOIN users u ON c.author_id = u.id
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortBy, sortOrder, argIndex, argIndex+1)

	args = append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search content: %w", err)
	}
	defer rows.Close()

	var items []domain.Content
	for rows.Next() {
		content := domain.Content{
			Tags:       []string{},
			Categories: []string{},
			Metadata:   make(map[string]interface{}),
		}

		var metadataJSON []byte
		var publishedAt sql.NullTime
		var mediaURL, mediaType sql.NullString
		var mediaSize sql.NullInt64

		err := rows.Scan(
			&content.ID, &content.Type, &content.Title, &content.Slug,
			&content.Body, &content.Excerpt, &content.Status, &content.AuthorID,
			pq.Array(&content.Tags), pq.Array(&content.Categories), &metadataJSON,
			&mediaURL, &mediaType, &mediaSize, &content.Version, &content.ViewCount,
			&publishedAt, &content.CreatedAt, &content.UpdatedAt, &content.AuthorName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan content: %w", err)
		}

		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &content.Metadata)
		}
		if publishedAt.Valid {
			content.PublishedAt = &publishedAt.Time
		}
		if mediaURL.Valid {
			content.MediaURL = mediaURL.String
		}
		if mediaType.Valid {
			content.MediaType = mediaType.String
		}
		if mediaSize.Valid {
			content.MediaSize = mediaSize.Int64
		}

		items = append(items, content)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating content: %w", err)
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &domain.PaginatedContent{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *contentRepository) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE content SET view_count = view_count + 1 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *contentRepository) GetVersions(ctx context.Context, contentID uuid.UUID) ([]domain.ContentVersion, error) {
	query := `
		SELECT id, content_id, version, title, body, excerpt, metadata,
		       changed_by, change_notes, created_at
		FROM content_versions
		WHERE content_id = $1
		ORDER BY version DESC
	`

	rows, err := r.db.QueryContext(ctx, query, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get versions: %w", err)
	}
	defer rows.Close()

	var versions []domain.ContentVersion
	for rows.Next() {
		var v domain.ContentVersion
		var metadataJSON []byte
		var changeNotes sql.NullString

		err := rows.Scan(
			&v.ID, &v.ContentID, &v.Version, &v.Title, &v.Body, &v.Excerpt,
			&metadataJSON, &v.ChangedBy, &changeNotes, &v.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}

		if len(metadataJSON) > 0 {
			v.Metadata = make(map[string]interface{})
			json.Unmarshal(metadataJSON, &v.Metadata)
		}
		if changeNotes.Valid {
			v.ChangeNotes = changeNotes.String
		}

		versions = append(versions, v)
	}

	return versions, nil
}

func (r *contentRepository) GetVersion(ctx context.Context, contentID uuid.UUID, version int) (*domain.ContentVersion, error) {
	query := `
		SELECT id, content_id, version, title, body, excerpt, metadata,
		       changed_by, change_notes, created_at
		FROM content_versions
		WHERE content_id = $1 AND version = $2
	`

	var v domain.ContentVersion
	var metadataJSON []byte
	var changeNotes sql.NullString

	err := r.db.QueryRowContext(ctx, query, contentID, version).Scan(
		&v.ID, &v.ContentID, &v.Version, &v.Title, &v.Body, &v.Excerpt,
		&metadataJSON, &v.ChangedBy, &changeNotes, &v.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrContentNotFound
		}
		return nil, fmt.Errorf("failed to get version: %w", err)
	}

	if len(metadataJSON) > 0 {
		v.Metadata = make(map[string]interface{})
		json.Unmarshal(metadataJSON, &v.Metadata)
	}
	if changeNotes.Valid {
		v.ChangeNotes = changeNotes.String
	}

	return &v, nil
}
