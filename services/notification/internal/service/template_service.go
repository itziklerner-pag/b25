package service

import (
	"context"

	"github.com/b25/services/notification/internal/config"
	"github.com/b25/services/notification/internal/repository"
	"github.com/b25/services/notification/internal/templates"
	"go.uber.org/zap"
)

// TemplateService handles template-related business logic
type TemplateService struct {
	cfg            *config.Config
	logger         *zap.Logger
	templateRepo   repository.TemplateRepository
	templateEngine *templates.TemplateEngine
}

// NewTemplateService creates a new template service
func NewTemplateService(
	cfg *config.Config,
	logger *zap.Logger,
	templateRepo repository.TemplateRepository,
	templateEngine *templates.TemplateEngine,
) *TemplateService {
	return &TemplateService{
		cfg:            cfg,
		logger:         logger,
		templateRepo:   templateRepo,
		templateEngine: templateEngine,
	}
}

// TODO: Implement template service methods
