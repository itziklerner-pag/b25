package models

import (
	"time"
)

// MediaType represents the type of media file
type MediaType string

const (
	MediaTypeImage MediaType = "image"
	MediaTypeVideo MediaType = "video"
	MediaTypeAudio MediaType = "audio"
	MediaTypeFile  MediaType = "file"
)

// MediaStatus represents the processing status of media
type MediaStatus string

const (
	MediaStatusPending    MediaStatus = "pending"
	MediaStatusProcessing MediaStatus = "processing"
	MediaStatusReady      MediaStatus = "ready"
	MediaStatusFailed     MediaStatus = "failed"
)

// Media represents a media file in the system
type Media struct {
	ID           string      `json:"id" db:"id"`
	UserID       string      `json:"user_id" db:"user_id"`
	OrgID        string      `json:"org_id,omitempty" db:"org_id"`
	FileName     string      `json:"file_name" db:"file_name"`
	OriginalName string      `json:"original_name" db:"original_name"`
	MimeType     string      `json:"mime_type" db:"mime_type"`
	MediaType    MediaType   `json:"media_type" db:"media_type"`
	Size         int64       `json:"size" db:"size"`
	Status       MediaStatus `json:"status" db:"status"`
	StoragePath  string      `json:"-" db:"storage_path"`
	PublicURL    string      `json:"public_url" db:"public_url"`
	CDNUrl       string      `json:"cdn_url,omitempty" db:"cdn_url"`
	Metadata     Metadata    `json:"metadata,omitempty" db:"metadata"`
	CreatedAt    time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time  `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Metadata contains additional information about the media
type Metadata struct {
	Width       int               `json:"width,omitempty"`
	Height      int               `json:"height,omitempty"`
	Duration    float64           `json:"duration,omitempty"`
	Format      string            `json:"format,omitempty"`
	Bitrate     int               `json:"bitrate,omitempty"`
	Thumbnails  []ThumbnailInfo   `json:"thumbnails,omitempty"`
	Variants    []VariantInfo     `json:"variants,omitempty"`
	ExifData    map[string]string `json:"exif_data,omitempty"`
	ProcessedAt *time.Time        `json:"processed_at,omitempty"`
}

// ThumbnailInfo contains information about generated thumbnails
type ThumbnailInfo struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// VariantInfo contains information about media variants (different resolutions, formats)
type VariantInfo struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Format  string `json:"format"`
	Width   int    `json:"width,omitempty"`
	Height  int    `json:"height,omitempty"`
	Bitrate string `json:"bitrate,omitempty"`
}

// MediaQuota tracks storage usage per user/organization
type MediaQuota struct {
	ID        string    `json:"id" db:"id"`
	EntityID  string    `json:"entity_id" db:"entity_id"` // user_id or org_id
	Type      string    `json:"type" db:"type"`           // "user" or "org"
	Used      int64     `json:"used" db:"used"`
	Limit     int64     `json:"limit" db:"limit"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UploadSession represents a multipart upload session
type UploadSession struct {
	ID           string     `json:"id" db:"id"`
	UserID       string     `json:"user_id" db:"user_id"`
	FileName     string     `json:"file_name" db:"file_name"`
	TotalSize    int64      `json:"total_size" db:"total_size"`
	ChunkSize    int        `json:"chunk_size" db:"chunk_size"`
	TotalChunks  int        `json:"total_chunks" db:"total_chunks"`
	UploadedKeys []string   `json:"uploaded_keys" db:"uploaded_keys"`
	Status       string     `json:"status" db:"status"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt    time.Time  `json:"expires_at" db:"expires_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty" db:"completed_at"`
}

// MediaFilter represents filters for querying media
type MediaFilter struct {
	UserID    string
	OrgID     string
	MediaType MediaType
	Status    MediaStatus
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}

// MediaStats represents statistics about media usage
type MediaStats struct {
	TotalFiles  int64              `json:"total_files"`
	TotalSize   int64              `json:"total_size"`
	ByType      map[MediaType]int  `json:"by_type"`
	ByStatus    map[MediaStatus]int `json:"by_status"`
	QuotaUsed   int64              `json:"quota_used"`
	QuotaLimit  int64              `json:"quota_limit"`
	QuotaPercent float64           `json:"quota_percent"`
}
