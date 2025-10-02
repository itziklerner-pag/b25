package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/yourorg/b25/services/media/internal/config"
)

// LocalStorage implements Storage interface using local filesystem
type LocalStorage struct {
	basePath string
}

// NewLocalStorage creates a new local storage backend
func NewLocalStorage(cfg config.LocalConfig) (*LocalStorage, error) {
	// Ensure base directory exists
	if err := os.MkdirAll(cfg.BasePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &LocalStorage{
		basePath: cfg.BasePath,
	}, nil
}

// Upload uploads a file to local storage
func (l *LocalStorage) Upload(path string, reader io.Reader, contentType string) (string, error) {
	fullPath := filepath.Join(l.basePath, path)

	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file
	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy data
	if _, err := io.Copy(file, reader); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return path, nil
}

// Download downloads a file from local storage
func (l *LocalStorage) Download(path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(l.basePath, path)

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

// Delete removes a file from local storage
func (l *LocalStorage) Delete(path string) error {
	fullPath := filepath.Join(l.basePath, path)

	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// Exists checks if a file exists in local storage
func (l *LocalStorage) Exists(path string) (bool, error) {
	fullPath := filepath.Join(l.basePath, path)

	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// GetURL returns the public URL for a file (for local storage, this is just the path)
func (l *LocalStorage) GetURL(path string) string {
	return fmt.Sprintf("/media/%s", path)
}

// GetSignedURL returns a signed URL (for local storage, same as GetURL)
func (l *LocalStorage) GetSignedURL(path string, expiry int) (string, error) {
	return l.GetURL(path), nil
}

// Copy copies a file within local storage
func (l *LocalStorage) Copy(sourcePath, destPath string) error {
	srcFullPath := filepath.Join(l.basePath, sourcePath)
	dstFullPath := filepath.Join(l.basePath, destPath)

	// Ensure destination directory exists
	dir := filepath.Dir(dstFullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Open source file
	src, err := os.Open(srcFullPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(dstFullPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy data
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

// GetSize returns the size of a file
func (l *LocalStorage) GetSize(path string) (int64, error) {
	fullPath := filepath.Join(l.basePath, path)

	info, err := os.Stat(fullPath)
	if err != nil {
		return 0, fmt.Errorf("failed to get file info: %w", err)
	}

	return info.Size(), nil
}
