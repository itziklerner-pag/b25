package storage

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yourorg/b25/services/media/internal/config"
)

func TestLocalStorage(t *testing.T) {
	// Create temporary directory for testing
	tempDir := filepath.Join(os.TempDir(), "media-test")
	defer os.RemoveAll(tempDir)

	cfg := config.LocalConfig{
		BasePath: tempDir,
	}

	storage, err := NewLocalStorage(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, storage)

	t.Run("Upload and Download", func(t *testing.T) {
		testData := []byte("test file content")
		testPath := "test/file.txt"

		// Upload
		path, err := storage.Upload(testPath, bytes.NewReader(testData), "text/plain")
		assert.NoError(t, err)
		assert.Equal(t, testPath, path)

		// Download
		reader, err := storage.Download(testPath)
		assert.NoError(t, err)
		defer reader.Close()

		downloaded, err := io.ReadAll(reader)
		assert.NoError(t, err)
		assert.Equal(t, testData, downloaded)
	})

	t.Run("Exists", func(t *testing.T) {
		testPath := "test/file.txt"

		exists, err := storage.Exists(testPath)
		assert.NoError(t, err)
		assert.True(t, exists)

		exists, err = storage.Exists("nonexistent/file.txt")
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Copy", func(t *testing.T) {
		srcPath := "test/file.txt"
		dstPath := "test/file-copy.txt"

		err := storage.Copy(srcPath, dstPath)
		assert.NoError(t, err)

		exists, err := storage.Exists(dstPath)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("GetSize", func(t *testing.T) {
		testPath := "test/file.txt"

		size, err := storage.GetSize(testPath)
		assert.NoError(t, err)
		assert.Greater(t, size, int64(0))
	})

	t.Run("Delete", func(t *testing.T) {
		testPath := "test/file-copy.txt"

		err := storage.Delete(testPath)
		assert.NoError(t, err)

		exists, err := storage.Exists(testPath)
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}
