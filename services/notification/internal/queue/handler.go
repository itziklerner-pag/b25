package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/b25/services/notification/internal/models"
	"github.com/b25/services/notification/internal/service"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// TaskHandler handles notification tasks from the queue
type TaskHandler struct {
	notificationService *service.NotificationService
	logger              *zap.Logger
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(notificationService *service.NotificationService, logger *zap.Logger) *TaskHandler {
	return &TaskHandler{
		notificationService: notificationService,
		logger:              logger,
	}
}

// HandleSendNotification handles the send notification task
func (h *TaskHandler) HandleSendNotification(ctx context.Context, task *asynq.Task) error {
	var notification models.Notification
	if err := json.Unmarshal(task.Payload(), &notification); err != nil {
		h.logger.Error("failed to unmarshal notification", zap.Error(err))
		return fmt.Errorf("failed to unmarshal notification: %w", err)
	}

	h.logger.Info("processing notification task",
		zap.String("notification_id", notification.ID.String()),
		zap.String("channel", string(notification.Channel)),
		zap.Int("retry_count", task.ResultWriter().(*asynq.ResultWriter).TaskInfo().Retried),
	)

	if err := h.notificationService.SendNotification(ctx, &notification); err != nil {
		h.logger.Error("failed to send notification",
			zap.String("notification_id", notification.ID.String()),
			zap.Error(err),
		)
		return err
	}

	h.logger.Info("notification sent successfully",
		zap.String("notification_id", notification.ID.String()),
	)

	return nil
}
