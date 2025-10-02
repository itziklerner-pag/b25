package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/b25/services/notification/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// NotificationRepository handles database operations for notifications
type NotificationRepository interface {
	// Notification CRUD
	Create(ctx context.Context, notification *models.Notification) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Notification, error)
	Update(ctx context.Context, notification *models.Notification) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter *models.QueryFilter) ([]*models.Notification, int, error)

	// Status updates
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.NotificationStatus) error
	MarkAsSent(ctx context.Context, id uuid.UUID, externalID string) error
	MarkAsDelivered(ctx context.Context, id uuid.UUID) error
	MarkAsFailed(ctx context.Context, id uuid.UUID, errorMsg, errorCode string) error

	// Retry management
	GetPendingRetries(ctx context.Context, limit int) ([]*models.Notification, error)
	UpdateRetryInfo(ctx context.Context, id uuid.UUID, retryCount int, nextRetryAt time.Time) error

	// Scheduled notifications
	GetScheduledNotifications(ctx context.Context, before time.Time, limit int) ([]*models.Notification, error)

	// User operations
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Notification, int, error)
	GetByCorrelationID(ctx context.Context, correlationID string) ([]*models.Notification, error)

	// Statistics
	GetStats(ctx context.Context, from, to time.Time) ([]*models.NotificationStats, error)
}

type notificationRepository struct {
	db *sqlx.DB
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *sqlx.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, notification *models.Notification) error {
	query := `
		INSERT INTO notifications (
			id, user_id, channel, template_id, priority, status,
			subject, body, metadata, recipient_address,
			scheduled_at, max_retries, correlation_id
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
		RETURNING created_at, updated_at
	`

	metadata, err := json.Marshal(notification.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	err = r.db.QueryRowContext(
		ctx, query,
		notification.ID,
		notification.UserID,
		notification.Channel,
		notification.TemplateID,
		notification.Priority,
		notification.Status,
		notification.Subject,
		notification.Body,
		metadata,
		notification.RecipientAddress,
		notification.ScheduledAt,
		notification.MaxRetries,
		notification.CorrelationID,
	).Scan(&notification.CreatedAt, &notification.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

func (r *notificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Notification, error) {
	query := `
		SELECT
			id, user_id, channel, template_id, priority, status,
			subject, body, metadata, recipient_address,
			scheduled_at, sent_at, delivered_at, failed_at,
			retry_count, max_retries, next_retry_at,
			error_message, error_code, external_message_id, correlation_id,
			created_at, updated_at
		FROM notifications
		WHERE id = $1
	`

	var notification models.Notification
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Channel,
		&notification.TemplateID,
		&notification.Priority,
		&notification.Status,
		&notification.Subject,
		&notification.Body,
		&metadataJSON,
		&notification.RecipientAddress,
		&notification.ScheduledAt,
		&notification.SentAt,
		&notification.DeliveredAt,
		&notification.FailedAt,
		&notification.RetryCount,
		&notification.MaxRetries,
		&notification.NextRetryAt,
		&notification.ErrorMessage,
		&notification.ErrorCode,
		&notification.ExternalMessageID,
		&notification.CorrelationID,
		&notification.CreatedAt,
		&notification.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("notification not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &notification.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &notification, nil
}

func (r *notificationRepository) Update(ctx context.Context, notification *models.Notification) error {
	query := `
		UPDATE notifications SET
			status = $2,
			subject = $3,
			body = $4,
			metadata = $5,
			recipient_address = $6,
			sent_at = $7,
			delivered_at = $8,
			failed_at = $9,
			retry_count = $10,
			next_retry_at = $11,
			error_message = $12,
			error_code = $13,
			external_message_id = $14
		WHERE id = $1
	`

	metadata, err := json.Marshal(notification.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	result, err := r.db.ExecContext(
		ctx, query,
		notification.ID,
		notification.Status,
		notification.Subject,
		notification.Body,
		metadata,
		notification.RecipientAddress,
		notification.SentAt,
		notification.DeliveredAt,
		notification.FailedAt,
		notification.RetryCount,
		notification.NextRetryAt,
		notification.ErrorMessage,
		notification.ErrorCode,
		notification.ExternalMessageID,
	)

	if err != nil {
		return fmt.Errorf("failed to update notification: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

func (r *notificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM notifications WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

func (r *notificationRepository) List(ctx context.Context, filter *models.QueryFilter) ([]*models.Notification, int, error) {
	// Build query with filters
	query := `
		SELECT
			id, user_id, channel, template_id, priority, status,
			subject, body, metadata, recipient_address,
			scheduled_at, sent_at, delivered_at, failed_at,
			retry_count, max_retries, next_retry_at,
			error_message, error_code, external_message_id, correlation_id,
			created_at, updated_at
		FROM notifications
		WHERE 1=1
	`
	countQuery := `SELECT COUNT(*) FROM notifications WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	// Apply filters
	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *filter.UserID)
		argIndex++
	}

	if filter.Channel != nil {
		query += fmt.Sprintf(" AND channel = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND channel = $%d", argIndex)
		args = append(args, *filter.Channel)
		argIndex++
	}

	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *filter.Status)
		argIndex++
	}

	if filter.Priority != nil {
		query += fmt.Sprintf(" AND priority = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND priority = $%d", argIndex)
		args = append(args, *filter.Priority)
		argIndex++
	}

	if filter.CorrelationID != nil {
		query += fmt.Sprintf(" AND correlation_id = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND correlation_id = $%d", argIndex)
		args = append(args, *filter.CorrelationID)
		argIndex++
	}

	if filter.StartDate != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		countQuery += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *filter.StartDate)
		argIndex++
	}

	if filter.EndDate != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		countQuery += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, *filter.EndDate)
		argIndex++
	}

	// Get total count
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	// Add ordering and pagination
	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
		argIndex++
	}

	// Execute query
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list notifications: %w", err)
	}
	defer rows.Close()

	notifications := make([]*models.Notification, 0)
	for rows.Next() {
		var notification models.Notification
		var metadataJSON []byte

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Channel,
			&notification.TemplateID,
			&notification.Priority,
			&notification.Status,
			&notification.Subject,
			&notification.Body,
			&metadataJSON,
			&notification.RecipientAddress,
			&notification.ScheduledAt,
			&notification.SentAt,
			&notification.DeliveredAt,
			&notification.FailedAt,
			&notification.RetryCount,
			&notification.MaxRetries,
			&notification.NextRetryAt,
			&notification.ErrorMessage,
			&notification.ErrorCode,
			&notification.ExternalMessageID,
			&notification.CorrelationID,
			&notification.CreatedAt,
			&notification.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan notification: %w", err)
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &notification.Metadata); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		notifications = append(notifications, &notification)
	}

	return notifications, total, nil
}

func (r *notificationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.NotificationStatus) error {
	query := `UPDATE notifications SET status = $2 WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

func (r *notificationRepository) MarkAsSent(ctx context.Context, id uuid.UUID, externalID string) error {
	query := `
		UPDATE notifications
		SET status = $2, sent_at = NOW(), external_message_id = $3
		WHERE id = $1
	`
	result, err := r.db.ExecContext(ctx, query, id, models.StatusSent, externalID)
	if err != nil {
		return fmt.Errorf("failed to mark as sent: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

func (r *notificationRepository) MarkAsDelivered(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE notifications
		SET status = $2, delivered_at = NOW()
		WHERE id = $1
	`
	result, err := r.db.ExecContext(ctx, query, id, models.StatusDelivered)
	if err != nil {
		return fmt.Errorf("failed to mark as delivered: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

func (r *notificationRepository) MarkAsFailed(ctx context.Context, id uuid.UUID, errorMsg, errorCode string) error {
	query := `
		UPDATE notifications
		SET status = $2, failed_at = NOW(), error_message = $3, error_code = $4
		WHERE id = $1
	`
	result, err := r.db.ExecContext(ctx, query, id, models.StatusFailed, errorMsg, errorCode)
	if err != nil {
		return fmt.Errorf("failed to mark as failed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

func (r *notificationRepository) GetPendingRetries(ctx context.Context, limit int) ([]*models.Notification, error) {
	query := `
		SELECT
			id, user_id, channel, template_id, priority, status,
			subject, body, metadata, recipient_address,
			scheduled_at, sent_at, delivered_at, failed_at,
			retry_count, max_retries, next_retry_at,
			error_message, error_code, external_message_id, correlation_id,
			created_at, updated_at
		FROM notifications
		WHERE status = $1
		  AND next_retry_at <= NOW()
		  AND retry_count < max_retries
		ORDER BY priority DESC, next_retry_at ASC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, models.StatusRetrying, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending retries: %w", err)
	}
	defer rows.Close()

	notifications := make([]*models.Notification, 0)
	for rows.Next() {
		var notification models.Notification
		var metadataJSON []byte

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Channel,
			&notification.TemplateID,
			&notification.Priority,
			&notification.Status,
			&notification.Subject,
			&notification.Body,
			&metadataJSON,
			&notification.RecipientAddress,
			&notification.ScheduledAt,
			&notification.SentAt,
			&notification.DeliveredAt,
			&notification.FailedAt,
			&notification.RetryCount,
			&notification.MaxRetries,
			&notification.NextRetryAt,
			&notification.ErrorMessage,
			&notification.ErrorCode,
			&notification.ExternalMessageID,
			&notification.CorrelationID,
			&notification.CreatedAt,
			&notification.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &notification.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		notifications = append(notifications, &notification)
	}

	return notifications, nil
}

func (r *notificationRepository) UpdateRetryInfo(ctx context.Context, id uuid.UUID, retryCount int, nextRetryAt time.Time) error {
	query := `
		UPDATE notifications
		SET retry_count = $2, next_retry_at = $3, status = $4
		WHERE id = $1
	`
	result, err := r.db.ExecContext(ctx, query, id, retryCount, nextRetryAt, models.StatusRetrying)
	if err != nil {
		return fmt.Errorf("failed to update retry info: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

func (r *notificationRepository) GetScheduledNotifications(ctx context.Context, before time.Time, limit int) ([]*models.Notification, error) {
	query := `
		SELECT
			id, user_id, channel, template_id, priority, status,
			subject, body, metadata, recipient_address,
			scheduled_at, sent_at, delivered_at, failed_at,
			retry_count, max_retries, next_retry_at,
			error_message, error_code, external_message_id, correlation_id,
			created_at, updated_at
		FROM notifications
		WHERE status = $1
		  AND scheduled_at IS NOT NULL
		  AND scheduled_at <= $2
		ORDER BY scheduled_at ASC, priority DESC
		LIMIT $3
	`

	rows, err := r.db.QueryContext(ctx, query, models.StatusPending, before, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduled notifications: %w", err)
	}
	defer rows.Close()

	notifications := make([]*models.Notification, 0)
	for rows.Next() {
		var notification models.Notification
		var metadataJSON []byte

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Channel,
			&notification.TemplateID,
			&notification.Priority,
			&notification.Status,
			&notification.Subject,
			&notification.Body,
			&metadataJSON,
			&notification.RecipientAddress,
			&notification.ScheduledAt,
			&notification.SentAt,
			&notification.DeliveredAt,
			&notification.FailedAt,
			&notification.RetryCount,
			&notification.MaxRetries,
			&notification.NextRetryAt,
			&notification.ErrorMessage,
			&notification.ErrorCode,
			&notification.ExternalMessageID,
			&notification.CorrelationID,
			&notification.CreatedAt,
			&notification.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &notification.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		notifications = append(notifications, &notification)
	}

	return notifications, nil
}

func (r *notificationRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Notification, int, error) {
	filter := &models.QueryFilter{
		UserID: &userID,
		Limit:  limit,
		Offset: offset,
	}
	return r.List(ctx, filter)
}

func (r *notificationRepository) GetByCorrelationID(ctx context.Context, correlationID string) ([]*models.Notification, error) {
	query := `
		SELECT
			id, user_id, channel, template_id, priority, status,
			subject, body, metadata, recipient_address,
			scheduled_at, sent_at, delivered_at, failed_at,
			retry_count, max_retries, next_retry_at,
			error_message, error_code, external_message_id, correlation_id,
			created_at, updated_at
		FROM notifications
		WHERE correlation_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, correlationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications by correlation ID: %w", err)
	}
	defer rows.Close()

	notifications := make([]*models.Notification, 0)
	for rows.Next() {
		var notification models.Notification
		var metadataJSON []byte

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Channel,
			&notification.TemplateID,
			&notification.Priority,
			&notification.Status,
			&notification.Subject,
			&notification.Body,
			&metadataJSON,
			&notification.RecipientAddress,
			&notification.ScheduledAt,
			&notification.SentAt,
			&notification.DeliveredAt,
			&notification.FailedAt,
			&notification.RetryCount,
			&notification.MaxRetries,
			&notification.NextRetryAt,
			&notification.ErrorMessage,
			&notification.ErrorCode,
			&notification.ExternalMessageID,
			&notification.CorrelationID,
			&notification.CreatedAt,
			&notification.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &notification.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		notifications = append(notifications, &notification)
	}

	return notifications, nil
}

func (r *notificationRepository) GetStats(ctx context.Context, from, to time.Time) ([]*models.NotificationStats, error) {
	query := `
		SELECT channel, status, priority, COUNT(*) as count, DATE_TRUNC('hour', created_at) as hour
		FROM notifications
		WHERE created_at BETWEEN $1 AND $2
		GROUP BY channel, status, priority, DATE_TRUNC('hour', created_at)
		ORDER BY hour DESC
	`

	rows, err := r.db.QueryContext(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	defer rows.Close()

	stats := make([]*models.NotificationStats, 0)
	for rows.Next() {
		var stat models.NotificationStats
		err := rows.Scan(&stat.Channel, &stat.Status, &stat.Priority, &stat.Count, &stat.Hour)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stats: %w", err)
		}
		stats = append(stats, &stat)
	}

	return stats, nil
}
