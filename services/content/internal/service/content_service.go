package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/b25/services/content/internal/domain"
	"github.com/b25/services/content/internal/repository"
	"github.com/google/uuid"
)

type ContentService interface {
	CreateContent(ctx context.Context, input domain.CreateContentInput, userID uuid.UUID) (*domain.Content, error)
	GetContent(ctx context.Context, id uuid.UUID) (*domain.Content, error)
	GetContentBySlug(ctx context.Context, slug string) (*domain.Content, error)
	UpdateContent(ctx context.Context, id uuid.UUID, input domain.UpdateContentInput, userID uuid.UUID) (*domain.Content, error)
	DeleteContent(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	SearchContent(ctx context.Context, params domain.SearchContentParams) (*domain.PaginatedContent, error)
	PublishContent(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Content, error)
	ArchiveContent(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Content, error)
	IncrementViewCount(ctx context.Context, id uuid.UUID) error
	GetContentVersions(ctx context.Context, contentID uuid.UUID) ([]domain.ContentVersion, error)
	GetContentVersion(ctx context.Context, contentID uuid.UUID, version int) (*domain.ContentVersion, error)
}

type contentService struct {
	contentRepo repository.ContentRepository
	userRepo    repository.UserRepository
}

func NewContentService(contentRepo repository.ContentRepository, userRepo repository.UserRepository) ContentService {
	return &contentService{
		contentRepo: contentRepo,
		userRepo:    userRepo,
	}
}

func (s *contentService) CreateContent(ctx context.Context, input domain.CreateContentInput, userID uuid.UUID) (*domain.Content, error) {
	// Validate input
	if err := input.Validate(); err != nil {
		return nil, domain.ErrInvalidInput
	}

	// Verify user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check user permissions
	if !user.HasPermission("create") {
		return nil, domain.ErrForbidden
	}

	// Generate slug if not provided
	slug := input.Slug
	if slug == "" {
		slug = generateSlug(input.Title)
	}

	// Create content entity
	content := &domain.Content{
		ID:         uuid.New(),
		Type:       input.Type,
		Title:      input.Title,
		Slug:       slug,
		Body:       input.Body,
		Excerpt:    input.Excerpt,
		Status:     input.Status,
		AuthorID:   userID,
		Tags:       input.Tags,
		Categories: input.Categories,
		Metadata:   input.Metadata,
		Version:    1,
		ViewCount:  0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Set published_at if status is published
	if content.Status == domain.ContentStatusPublished {
		now := time.Now()
		content.PublishedAt = &now
	}

	// Save to database
	if err := s.contentRepo.Create(ctx, content); err != nil {
		return nil, err
	}

	content.AuthorName = user.Username
	return content, nil
}

func (s *contentService) GetContent(ctx context.Context, id uuid.UUID) (*domain.Content, error) {
	content, err := s.contentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (s *contentService) GetContentBySlug(ctx context.Context, slug string) (*domain.Content, error) {
	content, err := s.contentRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (s *contentService) UpdateContent(ctx context.Context, id uuid.UUID, input domain.UpdateContentInput, userID uuid.UUID) (*domain.Content, error) {
	// Get existing content
	content, err := s.contentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check permissions
	if !user.CanEditContent(content) {
		return nil, domain.ErrForbidden
	}

	// Apply updates
	if input.Title != nil {
		content.Title = *input.Title
		// Regenerate slug if title changed
		if input.Slug == nil {
			content.Slug = generateSlug(*input.Title)
		}
	}
	if input.Slug != nil {
		content.Slug = *input.Slug
	}
	if input.Body != nil {
		content.Body = *input.Body
	}
	if input.Excerpt != nil {
		content.Excerpt = *input.Excerpt
	}
	if input.Status != nil {
		content.Status = *input.Status
		// Set published_at when publishing
		if *input.Status == domain.ContentStatusPublished && content.PublishedAt == nil {
			now := time.Now()
			content.PublishedAt = &now
		}
	}
	if len(input.Tags) > 0 {
		content.Tags = input.Tags
	}
	if len(input.Categories) > 0 {
		content.Categories = input.Categories
	}
	if input.Metadata != nil {
		content.Metadata = input.Metadata
	}

	// Update in database
	if err := s.contentRepo.Update(ctx, id, content); err != nil {
		return nil, err
	}

	return content, nil
}

func (s *contentService) DeleteContent(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// Get content
	content, err := s.contentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Check permissions (only admin or content owner can delete)
	if user.Role != domain.UserRoleAdmin && content.AuthorID != userID {
		return domain.ErrForbidden
	}

	return s.contentRepo.Delete(ctx, id)
}

func (s *contentService) SearchContent(ctx context.Context, params domain.SearchContentParams) (*domain.PaginatedContent, error) {
	return s.contentRepo.Search(ctx, params)
}

func (s *contentService) PublishContent(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Content, error) {
	status := domain.ContentStatusPublished
	return s.UpdateContent(ctx, id, domain.UpdateContentInput{
		Status: &status,
	}, userID)
}

func (s *contentService) ArchiveContent(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Content, error) {
	status := domain.ContentStatusArchived
	return s.UpdateContent(ctx, id, domain.UpdateContentInput{
		Status: &status,
	}, userID)
}

func (s *contentService) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	return s.contentRepo.IncrementViewCount(ctx, id)
}

func (s *contentService) GetContentVersions(ctx context.Context, contentID uuid.UUID) ([]domain.ContentVersion, error) {
	return s.contentRepo.GetVersions(ctx, contentID)
}

func (s *contentService) GetContentVersion(ctx context.Context, contentID uuid.UUID, version int) (*domain.ContentVersion, error) {
	return s.contentRepo.GetVersion(ctx, contentID, version)
}

// Helper function to generate URL-friendly slugs
func generateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove special characters
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	slug = result.String()
	// Remove multiple consecutive dashes
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	// Trim dashes from start and end
	slug = strings.Trim(slug, "-")
	// Add timestamp to ensure uniqueness
	slug = fmt.Sprintf("%s-%d", slug, time.Now().Unix())
	return slug
}
