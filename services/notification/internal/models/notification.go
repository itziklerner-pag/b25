package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// NotificationChannel represents the delivery channel
type NotificationChannel string

const (
	ChannelEmail   NotificationChannel = "email"
	ChannelSMS     NotificationChannel = "sms"
	ChannelPush    NotificationChannel = "push"
	ChannelWebhook NotificationChannel = "webhook"
)

// NotificationStatus represents the delivery status
type NotificationStatus string

const (
	StatusPending   NotificationStatus = "pending"
	StatusQueued    NotificationStatus = "queued"
	StatusSending   NotificationStatus = "sending"
	StatusSent      NotificationStatus = "sent"
	StatusDelivered NotificationStatus = "delivered"
	StatusFailed    NotificationStatus = "failed"
	StatusRetrying  NotificationStatus = "retrying"
	StatusCancelled NotificationStatus = "cancelled"
)

// NotificationPriority represents the urgency level
type NotificationPriority string

const (
	PriorityLow      NotificationPriority = "low"
	PriorityNormal   NotificationPriority = "normal"
	PriorityHigh     NotificationPriority = "high"
	PriorityCritical NotificationPriority = "critical"
)

// TemplateType represents the notification category
type TemplateType string

const (
	TypeTradingAlert      TemplateType = "trading_alert"
	TypeRiskViolation     TemplateType = "risk_violation"
	TypeOrderFill         TemplateType = "order_fill"
	TypeAccountUpdate     TemplateType = "account_update"
	TypeSystemNotification TemplateType = "system_notification"
	TypeCustom            TemplateType = "custom"
)

// Notification represents a notification record
type Notification struct {
	ID               uuid.UUID            `json:"id" db:"id"`
	UserID           uuid.UUID            `json:"user_id" db:"user_id"`
	Channel          NotificationChannel  `json:"channel" db:"channel"`
	TemplateID       *uuid.UUID           `json:"template_id,omitempty" db:"template_id"`
	Priority         NotificationPriority `json:"priority" db:"priority"`
	Status           NotificationStatus   `json:"status" db:"status"`
	Subject          *string              `json:"subject,omitempty" db:"subject"`
	Body             string               `json:"body" db:"body"`
	Metadata         map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	RecipientAddress *string              `json:"recipient_address,omitempty" db:"recipient_address"`
	ScheduledAt      *time.Time           `json:"scheduled_at,omitempty" db:"scheduled_at"`
	SentAt           *time.Time           `json:"sent_at,omitempty" db:"sent_at"`
	DeliveredAt      *time.Time           `json:"delivered_at,omitempty" db:"delivered_at"`
	FailedAt         *time.Time           `json:"failed_at,omitempty" db:"failed_at"`
	RetryCount       int                  `json:"retry_count" db:"retry_count"`
	MaxRetries       int                  `json:"max_retries" db:"max_retries"`
	NextRetryAt      *time.Time           `json:"next_retry_at,omitempty" db:"next_retry_at"`
	ErrorMessage     *string              `json:"error_message,omitempty" db:"error_message"`
	ErrorCode        *string              `json:"error_code,omitempty" db:"error_code"`
	ExternalMessageID *string             `json:"external_message_id,omitempty" db:"external_message_id"`
	CorrelationID    *string              `json:"correlation_id,omitempty" db:"correlation_id"`
	CreatedAt        time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at" db:"updated_at"`
}

// CreateNotificationRequest represents a request to create a notification
type CreateNotificationRequest struct {
	UserID           string                 `json:"user_id" validate:"required"`
	Channel          NotificationChannel    `json:"channel" validate:"required,oneof=email sms push webhook"`
	TemplateName     *string                `json:"template_name,omitempty"`
	Priority         NotificationPriority   `json:"priority" validate:"omitempty,oneof=low normal high critical"`
	Subject          *string                `json:"subject,omitempty"`
	Body             *string                `json:"body,omitempty"`
	TemplateData     map[string]interface{} `json:"template_data,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	RecipientAddress *string                `json:"recipient_address,omitempty"`
	ScheduledAt      *time.Time             `json:"scheduled_at,omitempty"`
	CorrelationID    *string                `json:"correlation_id,omitempty"`
}

// User represents a notification recipient
type User struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	ExternalUserID string     `json:"external_user_id" db:"external_user_id"`
	Email          *string    `json:"email,omitempty" db:"email"`
	PhoneNumber    *string    `json:"phone_number,omitempty" db:"phone_number"`
	Timezone       string     `json:"timezone" db:"timezone"`
	Language       string     `json:"language" db:"language"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// UserDevice represents a device for push notifications
type UserDevice struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	DeviceToken string     `json:"device_token" db:"device_token"`
	DeviceType  string     `json:"device_type" db:"device_type"`
	DeviceName  *string    `json:"device_name,omitempty" db:"device_name"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	LastUsedAt  time.Time  `json:"last_used_at" db:"last_used_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// NotificationPreference represents user preferences
type NotificationPreference struct {
	ID                uuid.UUID           `json:"id" db:"id"`
	UserID            uuid.UUID           `json:"user_id" db:"user_id"`
	Channel           NotificationChannel `json:"channel" db:"channel"`
	Category          string              `json:"category" db:"category"`
	IsEnabled         bool                `json:"is_enabled" db:"is_enabled"`
	QuietHoursEnabled bool                `json:"quiet_hours_enabled" db:"quiet_hours_enabled"`
	QuietHoursStart   *string             `json:"quiet_hours_start,omitempty" db:"quiet_hours_start"`
	QuietHoursEnd     *string             `json:"quiet_hours_end,omitempty" db:"quiet_hours_end"`
	CreatedAt         time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at" db:"updated_at"`
}

// NotificationTemplate represents a reusable template
type NotificationTemplate struct {
	ID           uuid.UUID           `json:"id" db:"id"`
	Name         string              `json:"name" db:"name"`
	Type         TemplateType        `json:"type" db:"type"`
	Channel      NotificationChannel `json:"channel" db:"channel"`
	Subject      *string             `json:"subject,omitempty" db:"subject"`
	BodyTemplate string              `json:"body_template" db:"body_template"`
	Variables    map[string]interface{} `json:"variables,omitempty" db:"variables"`
	IsActive     bool                `json:"is_active" db:"is_active"`
	Version      int                 `json:"version" db:"version"`
	CreatedBy    *string             `json:"created_by,omitempty" db:"created_by"`
	CreatedAt    time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at" db:"updated_at"`
}

// NotificationEvent represents a delivery tracking event
type NotificationEvent struct {
	ID              uuid.UUID              `json:"id" db:"id"`
	NotificationID  uuid.UUID              `json:"notification_id" db:"notification_id"`
	EventType       string                 `json:"event_type" db:"event_type"`
	EventData       map[string]interface{} `json:"event_data,omitempty" db:"event_data"`
	ProviderEventID *string                `json:"provider_event_id,omitempty" db:"provider_event_id"`
	OccurredAt      time.Time              `json:"occurred_at" db:"occurred_at"`
}

// NotificationBatch represents a bulk notification operation
type NotificationBatch struct {
	ID          uuid.UUID           `json:"id" db:"id"`
	Name        *string             `json:"name,omitempty" db:"name"`
	Description *string             `json:"description,omitempty" db:"description"`
	Channel     NotificationChannel `json:"channel" db:"channel"`
	TemplateID  *uuid.UUID          `json:"template_id,omitempty" db:"template_id"`
	TotalCount  int                 `json:"total_count" db:"total_count"`
	SentCount   int                 `json:"sent_count" db:"sent_count"`
	FailedCount int                 `json:"failed_count" db:"failed_count"`
	Status      string              `json:"status" db:"status"`
	StartedAt   *time.Time          `json:"started_at,omitempty" db:"started_at"`
	CompletedAt *time.Time          `json:"completed_at,omitempty" db:"completed_at"`
	CreatedBy   *string             `json:"created_by,omitempty" db:"created_by"`
	CreatedAt   time.Time           `json:"created_at" db:"created_at"`
}

// AlertRule represents an automated notification rule
type AlertRule struct {
	ID               uuid.UUID              `json:"id" db:"id"`
	Name             string                 `json:"name" db:"name"`
	Description      *string                `json:"description,omitempty" db:"description"`
	EventSource      string                 `json:"event_source" db:"event_source"`
	EventType        string                 `json:"event_type" db:"event_type"`
	Conditions       map[string]interface{} `json:"conditions,omitempty" db:"conditions"`
	TemplateID       *uuid.UUID             `json:"template_id,omitempty" db:"template_id"`
	Channels         []NotificationChannel  `json:"channels" db:"channels"`
	Priority         NotificationPriority   `json:"priority" db:"priority"`
	RecipientGroups  []string               `json:"recipient_groups,omitempty" db:"recipient_groups"`
	IsActive         bool                   `json:"is_active" db:"is_active"`
	CooldownMinutes  int                    `json:"cooldown_minutes" db:"cooldown_minutes"`
	LastTriggeredAt  *time.Time             `json:"last_triggered_at,omitempty" db:"last_triggered_at"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
}

// Webhook represents a webhook configuration
type Webhook struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	UserID          *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	Name            string     `json:"name" db:"name"`
	URL             string     `json:"url" db:"url"`
	Secret          *string    `json:"secret,omitempty" db:"secret"`
	EventTypes      []string   `json:"event_types" db:"event_types"`
	IsActive        bool       `json:"is_active" db:"is_active"`
	RetryOnFailure  bool       `json:"retry_on_failure" db:"retry_on_failure"`
	MaxRetries      int        `json:"max_retries" db:"max_retries"`
	TimeoutSeconds  int        `json:"timeout_seconds" db:"timeout_seconds"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// NotificationStats represents aggregated statistics
type NotificationStats struct {
	Channel         NotificationChannel  `json:"channel"`
	Status          NotificationStatus   `json:"status"`
	Priority        NotificationPriority `json:"priority"`
	Count           int                  `json:"count"`
	Hour            time.Time            `json:"hour"`
}

// DeliveryResult represents the result of a notification delivery attempt
type DeliveryResult struct {
	Success           bool      `json:"success"`
	ExternalMessageID *string   `json:"external_message_id,omitempty"`
	ErrorMessage      *string   `json:"error_message,omitempty"`
	ErrorCode         *string   `json:"error_code,omitempty"`
	DeliveredAt       time.Time `json:"delivered_at"`
	ShouldRetry       bool      `json:"should_retry"`
}

// QueryFilter represents common query filters
type QueryFilter struct {
	UserID        *uuid.UUID            `json:"user_id,omitempty"`
	Channel       *NotificationChannel  `json:"channel,omitempty"`
	Status        *NotificationStatus   `json:"status,omitempty"`
	Priority      *NotificationPriority `json:"priority,omitempty"`
	CorrelationID *string               `json:"correlation_id,omitempty"`
	StartDate     *time.Time            `json:"start_date,omitempty"`
	EndDate       *time.Time            `json:"end_date,omitempty"`
	Limit         int                   `json:"limit,omitempty"`
	Offset        int                   `json:"offset,omitempty"`
}

// PaginatedResult represents a paginated response
type PaginatedResult struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
	HasMore    bool        `json:"has_more"`
}
