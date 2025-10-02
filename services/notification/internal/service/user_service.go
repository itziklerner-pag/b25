package service

import (
	"context"

	"github.com/b25/services/notification/internal/config"
	"github.com/b25/services/notification/internal/repository"
	"go.uber.org/zap"
)

// UserService handles user-related business logic
type UserService struct {
	cfg      *config.Config
	logger   *zap.Logger
	userRepo repository.UserRepository
}

// NewUserService creates a new user service
func NewUserService(
	cfg *config.Config,
	logger *zap.Logger,
	userRepo repository.UserRepository,
) *UserService {
	return &UserService{
		cfg:      cfg,
		logger:   logger,
		userRepo: userRepo,
	}
}

// TODO: Implement user service methods
