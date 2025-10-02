package domain

import (
	"time"

	"github.com/google/uuid"
)

// ContentType represents different types of content
type ContentType string

const (
	ContentTypePost    ContentType = "post"
	ContentTypeArticle ContentType = "article"
	ContentTypeMedia   ContentType = "media"
)

// ContentStatus represents the publication status of content
type ContentStatus string

const (
	ContentStatusDraft     ContentStatus = "draft"
	ContentStatusPublished ContentStatus = "published"
	ContentStatusArchived  ContentStatus = "archived"
)

// Content represents a content item (post, article, or media)
type Content struct {
	ID          uuid.UUID              `json:"id"`
	Type        ContentType            `json:"type"`
	Title       string                 `json:"title"`
	Slug        string                 `json:"slug"`
	Body        string                 `json:"body,omitempty"`
	Excerpt     string                 `json:"excerpt,omitempty"`
	Status      ContentStatus          `json:"status"`
	AuthorID    uuid.UUID              `json:"author_id"`
	AuthorName  string                 `json:"author_name,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Categories  []string               `json:"categories,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	MediaURL    string                 `json:"media_url,omitempty"`
	MediaType   string                 `json:"media_type,omitempty"`
	MediaSize   int64                  `json:"media_size,omitempty"`
	Version     int                    `json:"version"`
	ViewCount   int64                  `json:"view_count"`
	PublishedAt *time.Time             `json:"published_at,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ContentVersion represents a historical version of content
type ContentVersion struct {
	ID          uuid.UUID              `json:"id"`
	ContentID   uuid.UUID              `json:"content_id"`
	Version     int                    `json:"version"`
	Title       string                 `json:"title"`
	Body        string                 `json:"body,omitempty"`
	Excerpt     string                 `json:"excerpt,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	ChangedBy   uuid.UUID              `json:"changed_by"`
	ChangeNotes string                 `json:"change_notes,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// CreateContentInput represents input for creating content
type CreateContentInput struct {
	Type       ContentType            `json:"type" validate:"required,oneof=post article media"`
	Title      string                 `json:"title" validate:"required,min=1,max=255"`
	Slug       string                 `json:"slug" validate:"omitempty,min=1,max=255"`
	Body       string                 `json:"body" validate:"omitempty"`
	Excerpt    string                 `json:"excerpt" validate:"omitempty,max=500"`
	Status     ContentStatus          `json:"status" validate:"omitempty,oneof=draft published archived"`
	AuthorID   uuid.UUID              `json:"author_id" validate:"required"`
	Tags       []string               `json:"tags" validate:"omitempty,dive,min=1,max=50"`
	Categories []string               `json:"categories" validate:"omitempty,dive,min=1,max=50"`
	Metadata   map[string]interface{} `json:"metadata" validate:"omitempty"`
}

// UpdateContentInput represents input for updating content
type UpdateContentInput struct {
	Title       *string                 `json:"title" validate:"omitempty,min=1,max=255"`
	Slug        *string                 `json:"slug" validate:"omitempty,min=1,max=255"`
	Body        *string                 `json:"body" validate:"omitempty"`
	Excerpt     *string                 `json:"excerpt" validate:"omitempty,max=500"`
	Status      *ContentStatus          `json:"status" validate:"omitempty,oneof=draft published archived"`
	Tags        []string                `json:"tags" validate:"omitempty,dive,min=1,max=50"`
	Categories  []string                `json:"categories" validate:"omitempty,dive,min=1,max=50"`
	Metadata    map[string]interface{}  `json:"metadata" validate:"omitempty"`
	ChangeNotes string                  `json:"change_notes" validate:"omitempty,max=500"`
}

// SearchContentParams represents search and filter parameters
type SearchContentParams struct {
	Query      string          `json:"query" form:"query"`
	Type       ContentType     `json:"type" form:"type"`
	Status     ContentStatus   `json:"status" form:"status"`
	AuthorID   *uuid.UUID      `json:"author_id" form:"author_id"`
	Tags       []string        `json:"tags" form:"tags"`
	Categories []string        `json:"categories" form:"categories"`
	FromDate   *time.Time      `json:"from_date" form:"from_date"`
	ToDate     *time.Time      `json:"to_date" form:"to_date"`
	SortBy     string          `json:"sort_by" form:"sort_by"`      // created_at, updated_at, view_count, title
	SortOrder  string          `json:"sort_order" form:"sort_order"` // asc, desc
	Page       int             `json:"page" form:"page"`
	PageSize   int             `json:"page_size" form:"page_size"`
}

// PaginatedContent represents paginated content results
type PaginatedContent struct {
	Items      []Content `json:"items"`
	Total      int64     `json:"total"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	TotalPages int       `json:"total_pages"`
}

// Validate validates CreateContentInput
func (input *CreateContentInput) Validate() error {
	if input.Type == "" {
		input.Type = ContentTypePost
	}
	if input.Status == "" {
		input.Status = ContentStatusDraft
	}
	return nil
}

// IsValidContentType checks if a content type is valid
func IsValidContentType(t ContentType) bool {
	switch t {
	case ContentTypePost, ContentTypeArticle, ContentTypeMedia:
		return true
	}
	return false
}

// IsValidContentStatus checks if a content status is valid
func IsValidContentStatus(s ContentStatus) bool {
	switch s {
	case ContentStatusDraft, ContentStatusPublished, ContentStatusArchived:
		return true
	}
	return false
}
