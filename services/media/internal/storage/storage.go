package storage

import (
	"fmt"
	"io"

	"github.com/yourorg/b25/services/media/internal/config"
)

// Storage defines the interface for storage backends
type Storage interface {
	// Upload uploads a file to storage
	Upload(path string, reader io.Reader, contentType string) (string, error)

	// Download downloads a file from storage
	Download(path string) (io.ReadCloser, error)

	// Delete removes a file from storage
	Delete(path string) error

	// Exists checks if a file exists in storage
	Exists(path string) (bool, error)

	// GetURL returns the public URL for a file
	GetURL(path string) string

	// GetSignedURL returns a signed URL for temporary access
	GetSignedURL(path string, expiry int) (string, error)

	// Copy copies a file within storage
	Copy(sourcePath, destPath string) error

	// GetSize returns the size of a file
	GetSize(path string) (int64, error)
}

// NewStorage creates a new storage backend based on configuration
func NewStorage(cfg config.StorageConfig) (Storage, error) {
	switch cfg.Type {
	case "s3":
		return NewS3Storage(cfg.S3)
	case "local":
		return NewLocalStorage(cfg.Local)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", cfg.Type)
	}
}
