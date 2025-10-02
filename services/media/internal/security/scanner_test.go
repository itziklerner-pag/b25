package security

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeFileName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"normal-file.txt", "normal-file.txt"},
		{"../../../etc/passwd", ".._.._.._etc_passwd"},
		{"file with spaces.txt", "file with spaces.txt"},
		{"file\x00null.txt", "filenull.txt"},
		{"..hidden", "hidden"},
		{"  .leading-trailing.  ", "leading-trailing"},
		{"", "unnamed"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := SanitizeFileName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetFileExtension(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"file.txt", "txt"},
		{"image.jpeg", "jpeg"},
		{"archive.tar.gz", "gz"},
		{"no-extension", ""},
		{"multiple.dots.file.pdf", "pdf"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := GetFileExtension(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsImageFile(t *testing.T) {
	tests := []struct {
		mimeType string
		expected bool
	}{
		{"image/jpeg", true},
		{"image/png", true},
		{"image/gif", true},
		{"video/mp4", false},
		{"application/pdf", false},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			result := IsImageFile(tt.mimeType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsVideoFile(t *testing.T) {
	tests := []struct {
		mimeType string
		expected bool
	}{
		{"video/mp4", true},
		{"video/webm", true},
		{"image/jpeg", false},
		{"audio/mp3", false},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			result := IsVideoFile(tt.mimeType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNoOpScanner(t *testing.T) {
	scanner := NewNoOpScanner()
	assert.NotNil(t, scanner)

	clean, err := scanner.Scan(bytes.NewReader([]byte("test data")))
	assert.NoError(t, err)
	assert.True(t, clean)

	assert.True(t, scanner.IsAvailable())
}
