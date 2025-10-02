package providers

import (
	"context"

	"github.com/b25/services/notification/internal/models"
)

// Provider defines the interface for notification delivery providers
type Provider interface {
	// Send sends a notification and returns the result
	Send(ctx context.Context, notification *models.Notification) (*models.DeliveryResult, error)

	// GetChannel returns the channel this provider supports
	GetChannel() models.NotificationChannel

	// IsHealthy checks if the provider is healthy
	IsHealthy(ctx context.Context) error
}

// EmailProvider sends email notifications
type EmailProvider interface {
	Provider
	SendEmail(ctx context.Context, to, subject, body string, isHTML bool) (*models.DeliveryResult, error)
}

// SMSProvider sends SMS notifications
type SMSProvider interface {
	Provider
	SendSMS(ctx context.Context, to, message string) (*models.DeliveryResult, error)
}

// PushProvider sends push notifications
type PushProvider interface {
	Provider
	SendPush(ctx context.Context, deviceToken, title, body string, data map[string]interface{}) (*models.DeliveryResult, error)
}

// WebhookProvider sends webhook notifications
type WebhookProvider interface {
	Provider
	SendWebhook(ctx context.Context, url string, payload interface{}, secret *string) (*models.DeliveryResult, error)
}
