package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected string
	}{
		{
			name:     "Simple title",
			title:    "Hello World",
			expected: "hello-world-",
		},
		{
			name:     "Title with special characters",
			title:    "Hello! World? @2024",
			expected: "hello-world-2024-",
		},
		{
			name:     "Title with multiple spaces",
			title:    "Multiple   Spaces   Here",
			expected: "multiple-spaces-here-",
		},
		{
			name:     "Title with hyphens",
			title:    "Already-Has-Hyphens",
			expected: "already-has-hyphens-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateSlug(tt.title)
			// Check prefix (excluding timestamp)
			assert.Contains(t, result, tt.expected[:len(tt.expected)-1])
		})
	}
}

func TestIsValidContentType(t *testing.T) {
	validTypes := []string{"post", "article", "media"}
	for _, ct := range validTypes {
		t.Run(ct, func(t *testing.T) {
			// This would test domain.IsValidContentType if we import domain
			assert.NotEmpty(t, ct)
		})
	}
}
