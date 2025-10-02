package service

import (
	"context"
	"fmt"
	"time"

	"github.com/b25/services/notification/internal/config"
	"github.com/b25/services/notification/internal/models"
	"github.com/b25/services/notification/internal/providers"
	"github.com/b25/services/notification/internal/queue"
	"github.com/b25/services/notification/internal/repository"
	"github.com/b25/services/notification/internal/templates"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// NotificationService handles notification business logic
type NotificationService struct {
	cfg              *config.Config
	logger           *zap.Logger
	notifRepo        repository.NotificationRepository
	userRepo         repository.UserRepository
	templateRepo     repository.TemplateRepository
	templateEngine   *templates.TemplateEngine
	queue            queue.Queue
	providers        map[models.NotificationChannel]providers.Provider
	rateLimiter      RateLimiter
}

// NewNotificationService creates a new notification service
func NewNotificationService(
	cfg *config.Config,
	logger *zap.Logger,
	notifRepo repository.NotificationRepository,
	userRepo repository.UserRepository,
	templateRepo repository.TemplateRepository,
	templateEngine *templates.TemplateEngine,
	queue queue.Queue,
	providers map[models.NotificationChannel]providers.Provider,
	rateLimiter RateLimiter,
) *NotificationService {
	return &NotificationService{
		cfg:            cfg,
		logger:         logger,
		notifRepo:      notifRepo,
		userRepo:       userRepo,
		templateRepo:   templateRepo,
		templateEngine: templateEngine,
		queue:          queue,
		providers:      providers,
		rateLimiter:    rateLimiter,
	}
}

// CreateNotification creates and queues a notification
func (s *NotificationService) CreateNotification(ctx context.Context, req *models.CreateNotificationRequest) (*models.Notification, error) {
	// Validate request
	if err := s.validateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Get or create user
	user, err := s.userRepo.GetByExternalID(ctx, req.UserID)
	if err != nil {
		// User doesn't exist, return error (users should be created separately)
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Check rate limits
	if err := s.rateLimiter.CheckLimit(ctx, user.ID, req.Channel); err != nil {
		s.logger.Warn("rate limit exceeded",
			zap.String("user_id", user.ID.String()),
			zap.String("channel", string(req.Channel)),
		)
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Check user preferences
	if !s.shouldSendNotification(ctx, user.ID, req.Channel, req.Priority) {
		s.logger.Info("notification blocked by user preferences",
			zap.String("user_id", user.ID.String()),
			zap.String("channel", string(req.Channel)),
		)
		return nil, fmt.Errorf("notification blocked by user preferences")
	}

	// Create notification object
	notification := &models.Notification{
		ID:               uuid.New(),
		UserID:           user.ID,
		Channel:          req.Channel,
		Priority:         models.PriorityNormal,
		Status:           models.StatusPending,
		Body:             "",
		Metadata:         req.Metadata,
		RecipientAddress: req.RecipientAddress,
		ScheduledAt:      req.ScheduledAt,
		MaxRetries:       s.cfg.Queue.MaxRetry,
		CorrelationID:    req.CorrelationID,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if req.Priority != "" {
		notification.Priority = req.Priority
	}

	// Handle template-based notifications
	if req.TemplateName != nil {
		if err := s.applyTemplate(ctx, notification, *req.TemplateName, req.TemplateData); err != nil {
			return nil, fmt.Errorf("failed to apply template: %w", err)
		}
	} else {
		// Direct notification
		if req.Body == nil {
			return nil, fmt.Errorf("body is required when not using a template")
		}
		notification.Body = *req.Body
		notification.Subject = req.Subject
	}

	// Set recipient address if not provided
	if notification.RecipientAddress == nil {
		addr, err := s.getRecipientAddress(user, req.Channel)
		if err != nil {
			return nil, fmt.Errorf("failed to get recipient address: %w", err)
		}
		notification.RecipientAddress = &addr
	}

	// Save to database
	if err := s.notifRepo.Create(ctx, notification); err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	// Queue for delivery (unless scheduled)
	if notification.ScheduledAt == nil || notification.ScheduledAt.Before(time.Now()) {
		if err := s.queue.Enqueue(ctx, notification); err != nil {
			s.logger.Error("failed to queue notification",
				zap.String("notification_id", notification.ID.String()),
				zap.Error(err),
			)
			// Update status to failed
			notification.Status = models.StatusFailed
			s.notifRepo.Update(ctx, notification)
			return nil, fmt.Errorf("failed to queue notification: %w", err)
		}
		notification.Status = models.StatusQueued
		s.notifRepo.UpdateStatus(ctx, notification.ID, models.StatusQueued)
	}

	s.logger.Info("notification created",
		zap.String("notification_id", notification.ID.String()),
		zap.String("user_id", user.ID.String()),
		zap.String("channel", string(req.Channel)),
		zap.String("priority", string(notification.Priority)),
	)

	return notification, nil
}

// SendNotification sends a notification immediately
func (s *NotificationService) SendNotification(ctx context.Context, notification *models.Notification) error {
	s.logger.Info("sending notification",
		zap.String("notification_id", notification.ID.String()),
		zap.String("channel", string(notification.Channel)),
	)

	// Get provider for channel
	provider, exists := s.providers[notification.Channel]
	if !exists {
		return fmt.Errorf("no provider available for channel: %s", notification.Channel)
	}

	// Update status to sending
	if err := s.notifRepo.UpdateStatus(ctx, notification.ID, models.StatusSending); err != nil {
		s.logger.Error("failed to update status to sending", zap.Error(err))
	}

	// Send notification
	result, err := provider.Send(ctx, notification)
	if err != nil {
		s.logger.Error("failed to send notification",
			zap.String("notification_id", notification.ID.String()),
			zap.Error(err),
		)

		// Mark as failed or retry
		if result != nil && result.ShouldRetry && notification.RetryCount < notification.MaxRetries {
			return s.scheduleRetry(ctx, notification)
		}

		// Mark as permanently failed
		errorMsg := err.Error()
		errorCode := "SEND_ERROR"
		if result != nil && result.ErrorCode != nil {
			errorCode = *result.ErrorCode
		}
		if err := s.notifRepo.MarkAsFailed(ctx, notification.ID, errorMsg, errorCode); err != nil {
			s.logger.Error("failed to mark notification as failed", zap.Error(err))
		}

		return err
	}

	// Mark as sent/delivered
	if result.Success {
		externalID := ""
		if result.ExternalMessageID != nil {
			externalID = *result.ExternalMessageID
		}
		if err := s.notifRepo.MarkAsSent(ctx, notification.ID, externalID); err != nil {
			s.logger.Error("failed to mark notification as sent", zap.Error(err))
		}

		// For some channels, mark as delivered immediately
		if notification.Channel == models.ChannelSMS || notification.Channel == models.ChannelPush {
			if err := s.notifRepo.MarkAsDelivered(ctx, notification.ID); err != nil {
				s.logger.Error("failed to mark notification as delivered", zap.Error(err))
			}
		}
	}

	s.logger.Info("notification sent successfully",
		zap.String("notification_id", notification.ID.String()),
		zap.String("external_id", *result.ExternalMessageID),
	)

	return nil
}

// GetNotification retrieves a notification by ID
func (s *NotificationService) GetNotification(ctx context.Context, id uuid.UUID) (*models.Notification, error) {
	return s.notifRepo.GetByID(ctx, id)
}

// ListNotifications lists notifications with filters
func (s *NotificationService) ListNotifications(ctx context.Context, filter *models.QueryFilter) (*models.PaginatedResult, error) {
	notifications, total, err := s.notifRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	hasMore := false
	if filter.Limit > 0 && filter.Offset+filter.Limit < total {
		hasMore = true
	}

	return &models.PaginatedResult{
		Data:    notifications,
		Total:   total,
		Limit:   filter.Limit,
		Offset:  filter.Offset,
		HasMore: hasMore,
	}, nil
}

// GetUserNotifications retrieves notifications for a user
func (s *NotificationService) GetUserNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) (*models.PaginatedResult, error) {
	notifications, total, err := s.notifRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	hasMore := false
	if limit > 0 && offset+limit < total {
		hasMore = true
	}

	return &models.PaginatedResult{
		Data:    notifications,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		HasMore: hasMore,
	}, nil
}

// ProcessScheduledNotifications processes notifications scheduled for delivery
func (s *NotificationService) ProcessScheduledNotifications(ctx context.Context) error {
	notifications, err := s.notifRepo.GetScheduledNotifications(ctx, time.Now(), 100)
	if err != nil {
		return fmt.Errorf("failed to get scheduled notifications: %w", err)
	}

	s.logger.Info("processing scheduled notifications", zap.Int("count", len(notifications)))

	for _, notification := range notifications {
		if err := s.queue.Enqueue(ctx, notification); err != nil {
			s.logger.Error("failed to queue scheduled notification",
				zap.String("notification_id", notification.ID.String()),
				zap.Error(err),
			)
			continue
		}

		if err := s.notifRepo.UpdateStatus(ctx, notification.ID, models.StatusQueued); err != nil {
			s.logger.Error("failed to update notification status",
				zap.String("notification_id", notification.ID.String()),
				zap.Error(err),
			)
		}
	}

	return nil
}

// ProcessRetries processes notifications pending retry
func (s *NotificationService) ProcessRetries(ctx context.Context) error {
	notifications, err := s.notifRepo.GetPendingRetries(ctx, 100)
	if err != nil {
		return fmt.Errorf("failed to get pending retries: %w", err)
	}

	s.logger.Info("processing retry notifications", zap.Int("count", len(notifications)))

	for _, notification := range notifications {
		if err := s.queue.Enqueue(ctx, notification); err != nil {
			s.logger.Error("failed to queue retry notification",
				zap.String("notification_id", notification.ID.String()),
				zap.Error(err),
			)
			continue
		}
	}

	return nil
}

// Helper methods

func (s *NotificationService) validateRequest(req *models.CreateNotificationRequest) error {
	if req.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if req.Channel == "" {
		return fmt.Errorf("channel is required")
	}
	if req.TemplateName == nil && req.Body == nil {
		return fmt.Errorf("either template_name or body is required")
	}
	return nil
}

func (s *NotificationService) shouldSendNotification(ctx context.Context, userID uuid.UUID, channel models.NotificationChannel, priority models.NotificationPriority) bool {
	// Always send critical notifications
	if priority == models.PriorityCritical {
		return true
	}

	// Check user preferences
	// For now, default to allowing all notifications
	// TODO: Implement actual preference checking with quiet hours
	return true
}

func (s *NotificationService) applyTemplate(ctx context.Context, notification *models.Notification, templateName string, data map[string]interface{}) error {
	// Get template from database
	template, err := s.templateRepo.GetByName(ctx, templateName)
	if err != nil {
		return fmt.Errorf("template not found: %w", err)
	}

	notification.TemplateID = &template.ID

	// Render template
	body, err := s.templateEngine.Render(templateName, data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	notification.Body = body

	// Render subject if exists
	if template.Subject != nil && *template.Subject != "" {
		subject := *template.Subject
		notification.Subject = &subject
	}

	return nil
}

func (s *NotificationService) getRecipientAddress(user *models.User, channel models.NotificationChannel) (string, error) {
	switch channel {
	case models.ChannelEmail:
		if user.Email == nil {
			return "", fmt.Errorf("user has no email address")
		}
		return *user.Email, nil
	case models.ChannelSMS:
		if user.PhoneNumber == nil {
			return "", fmt.Errorf("user has no phone number")
		}
		return *user.PhoneNumber, nil
	case models.ChannelPush:
		// For push, we'll need to get an active device token
		// This is simplified; real implementation would query user_devices table
		return "", fmt.Errorf("push notifications require explicit device token")
	default:
		return "", fmt.Errorf("unsupported channel: %s", channel)
	}
}

func (s *NotificationService) scheduleRetry(ctx context.Context, notification *models.Notification) error {
	retryCount := notification.RetryCount + 1

	// Calculate exponential backoff
	delaySeconds := s.cfg.Queue.RetryDelayBase.Seconds() * float64(1<<uint(retryCount))
	nextRetryAt := time.Now().Add(time.Duration(delaySeconds) * time.Second)

	s.logger.Info("scheduling retry",
		zap.String("notification_id", notification.ID.String()),
		zap.Int("retry_count", retryCount),
		zap.Time("next_retry_at", nextRetryAt),
	)

	return s.notifRepo.UpdateRetryInfo(ctx, notification.ID, retryCount, nextRetryAt)
}
