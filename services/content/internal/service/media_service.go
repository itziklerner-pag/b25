package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/b25/services/content/internal/domain"
	"github.com/b25/services/content/internal/repository"
	"github.com/google/uuid"
)

const (
	MaxFileSize        = 50 * 1024 * 1024 // 50 MB
	DefaultUploadPath  = "./uploads"
)

var AllowedMediaTypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/gif":       true,
	"image/webp":      true,
	"video/mp4":       true,
	"video/webm":      true,
	"audio/mpeg":      true,
	"audio/mp3":       true,
	"audio/wav":       true,
	"application/pdf": true,
}

type MediaService interface {
	UploadMedia(ctx context.Context, file *multipart.FileHeader, userID uuid.UUID) (*domain.Content, error)
	DeleteMedia(ctx context.Context, contentID uuid.UUID, userID uuid.UUID) error
}

type mediaService struct {
	contentRepo repository.ContentRepository
	userRepo    repository.UserRepository
	uploadPath  string
	baseURL     string
}

func NewMediaService(
	contentRepo repository.ContentRepository,
	userRepo repository.UserRepository,
	uploadPath string,
	baseURL string,
) MediaService {
	if uploadPath == "" {
		uploadPath = DefaultUploadPath
	}
	return &mediaService{
		contentRepo: contentRepo,
		userRepo:    userRepo,
		uploadPath:  uploadPath,
		baseURL:     baseURL,
	}
}

func (s *mediaService) UploadMedia(ctx context.Context, file *multipart.FileHeader, userID uuid.UUID) (*domain.Content, error) {
	// Verify user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check user permissions
	if !user.HasPermission("create") {
		return nil, domain.ErrForbidden
	}

	// Validate file size
	if file.Size > MaxFileSize {
		return nil, domain.ErrMediaTooLarge
	}

	// Validate media type
	mediaType := file.Header.Get("Content-Type")
	if mediaType == "" {
		mediaType = detectContentType(file.Filename)
	}
	if !AllowedMediaTypes[mediaType] {
		return nil, domain.ErrInvalidMediaType
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// Create upload directory structure: uploads/YYYY/MM/DD/
	now := time.Now()
	dateFolder := now.Format("2006/01/02")
	uploadDir := filepath.Join(s.uploadPath, dateFolder)

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Full file path
	filePath := filepath.Join(uploadDir, filename)

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy file
	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(filePath) // Clean up on error
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Generate media URL
	mediaURL := fmt.Sprintf("%s/%s/%s", s.baseURL, dateFolder, filename)

	// Create content entity
	title := strings.TrimSuffix(file.Filename, ext)
	slug := generateSlug(title)

	content := &domain.Content{
		ID:         uuid.New(),
		Type:       domain.ContentTypeMedia,
		Title:      title,
		Slug:       slug,
		Status:     domain.ContentStatusPublished,
		AuthorID:   userID,
		MediaURL:   mediaURL,
		MediaType:  mediaType,
		MediaSize:  file.Size,
		Version:    1,
		ViewCount:  0,
		Tags:       []string{"media"},
		Categories: []string{},
		Metadata:   map[string]interface{}{
			"original_filename": file.Filename,
			"file_path":         filePath,
			"upload_date":       now.Format(time.RFC3339),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	publishedAt := now
	content.PublishedAt = &publishedAt

	// Save to database
	if err := s.contentRepo.Create(ctx, content); err != nil {
		os.Remove(filePath) // Clean up file on database error
		return nil, err
	}

	content.AuthorName = user.Username
	return content, nil
}

func (s *mediaService) DeleteMedia(ctx context.Context, contentID uuid.UUID, userID uuid.UUID) error {
	// Get content
	content, err := s.contentRepo.GetByID(ctx, contentID)
	if err != nil {
		return err
	}

	// Verify it's media type
	if content.Type != domain.ContentTypeMedia {
		return domain.ErrInvalidMediaType
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Check permissions
	if user.Role != domain.UserRoleAdmin && content.AuthorID != userID {
		return domain.ErrForbidden
	}

	// Delete file from filesystem
	if filePath, ok := content.Metadata["file_path"].(string); ok {
		os.Remove(filePath) // Ignore error if file doesn't exist
	}

	// Delete from database
	return s.contentRepo.Delete(ctx, contentID)
}

func detectContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".mp3":
		return "audio/mp3"
	case ".wav":
		return "audio/wav"
	case ".pdf":
		return "application/pdf"
	default:
		return "application/octet-stream"
	}
}
