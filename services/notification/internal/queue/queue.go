package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/b25/services/notification/internal/config"
	"github.com/b25/services/notification/internal/models"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

const (
	TypeSendNotification = "notification:send"
)

// Queue interface for notification queue operations
type Queue interface {
	Enqueue(ctx context.Context, notification *models.Notification) error
	EnqueueWithDelay(ctx context.Context, notification *models.Notification, delay time.Duration) error
	Close() error
}

// AsynqQueue implements queue using Asynq
type AsynqQueue struct {
	client *asynq.Client
	logger *zap.Logger
	cfg    *config.QueueConfig
}

// NewAsynqQueue creates a new Asynq-based queue
func NewAsynqQueue(cfg *config.QueueConfig, logger *zap.Logger) (Queue, error) {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr: cfg.RedisAddr,
	})

	return &AsynqQueue{
		client: client,
		logger: logger,
		cfg:    cfg,
	}, nil
}

// Enqueue adds a notification to the queue
func (q *AsynqQueue) Enqueue(ctx context.Context, notification *models.Notification) error {
	return q.EnqueueWithDelay(ctx, notification, 0)
}

// EnqueueWithDelay adds a notification to the queue with a delay
func (q *AsynqQueue) EnqueueWithDelay(ctx context.Context, notification *models.Notification, delay time.Duration) error {
	payload, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	task := asynq.NewTask(TypeSendNotification, payload)

	opts := []asynq.Option{
		asynq.MaxRetry(q.cfg.MaxRetry),
		asynq.Retention(time.Duration(q.cfg.RetentionDays) * 24 * time.Hour),
		asynq.Queue(string(notification.Priority)),
	}

	if delay > 0 {
		opts = append(opts, asynq.ProcessIn(delay))
	}

	// Set timeout based on channel
	timeout := 30 * time.Second
	if notification.Channel == models.ChannelEmail {
		timeout = 60 * time.Second
	}
	opts = append(opts, asynq.Timeout(timeout))

	info, err := q.client.EnqueueContext(ctx, task, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	q.logger.Debug("notification enqueued",
		zap.String("task_id", info.ID),
		zap.String("notification_id", notification.ID.String()),
		zap.String("queue", info.Queue),
	)

	return nil
}

// Close closes the queue client
func (q *AsynqQueue) Close() error {
	return q.client.Close()
}

// Worker handles the processing of queued notifications
type Worker struct {
	server  *asynq.Server
	mux     *asynq.ServeMux
	logger  *zap.Logger
	handler NotificationHandler
}

// NotificationHandler processes notification tasks
type NotificationHandler interface {
	HandleSendNotification(ctx context.Context, task *asynq.Task) error
}

// NewWorker creates a new queue worker
func NewWorker(cfg *config.QueueConfig, logger *zap.Logger, handler NotificationHandler) *Worker {
	server := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.RedisAddr},
		asynq.Config{
			Concurrency: cfg.Concurrency,
			Queues: map[string]int{
				string(models.PriorityCritical): 6,
				string(models.PriorityHigh):     4,
				string(models.PriorityNormal):   2,
				string(models.PriorityLow):      1,
			},
			RetryDelayFunc: func(n int, err error, task *asynq.Task) time.Duration {
				// Exponential backoff
				delay := cfg.RetryDelayBase * time.Duration(1<<uint(n))
				maxDelay := 1 * time.Hour
				if delay > maxDelay {
					delay = maxDelay
				}
				return delay
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				logger.Error("task failed",
					zap.String("type", task.Type()),
					zap.Error(err),
				)
			}),
		},
	)

	mux := asynq.NewServeMux()

	return &Worker{
		server:  server,
		mux:     mux,
		logger:  logger,
		handler: handler,
	}
}

// RegisterHandlers registers task handlers
func (w *Worker) RegisterHandlers() {
	w.mux.HandleFunc(TypeSendNotification, w.handler.HandleSendNotification)
}

// Start starts the worker
func (w *Worker) Start() error {
	w.logger.Info("starting notification worker")
	return w.server.Start(w.mux)
}

// Stop stops the worker gracefully
func (w *Worker) Stop() {
	w.logger.Info("stopping notification worker")
	w.server.Stop()
	w.server.Shutdown()
}
