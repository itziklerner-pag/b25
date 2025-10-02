package providers

import (
	"context"
	"fmt"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/b25/services/notification/internal/config"
	"github.com/b25/services/notification/internal/models"
	"google.golang.org/api/option"
)

// FCMProvider implements push notifications using Firebase Cloud Messaging
type FCMProvider struct {
	client *messaging.Client
}

// NewFCMProvider creates a new FCM provider
func NewFCMProvider(ctx context.Context, cfg *config.PushConfig) (PushProvider, error) {
	if cfg.FCM.CredentialsFile == "" {
		return nil, fmt.Errorf("FCM credentials file is required")
	}

	opt := option.WithCredentialsFile(cfg.FCM.CredentialsFile)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %w", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize FCM client: %w", err)
	}

	return &FCMProvider{
		client: client,
	}, nil
}

// Send sends a push notification
func (p *FCMProvider) Send(ctx context.Context, notification *models.Notification) (*models.DeliveryResult, error) {
	if notification.RecipientAddress == nil {
		return nil, fmt.Errorf("device token is required for push notification")
	}

	title := ""
	if notification.Subject != nil {
		title = *notification.Subject
	}

	// Extract data from metadata
	data := make(map[string]interface{})
	if notification.Metadata != nil {
		data = notification.Metadata
	}

	return p.SendPush(ctx, *notification.RecipientAddress, title, notification.Body, data)
}

// SendPush sends a push notification using FCM
func (p *FCMProvider) SendPush(ctx context.Context, deviceToken, title, body string, data map[string]interface{}) (*models.DeliveryResult, error) {
	// Convert data map to string map for FCM
	dataStr := make(map[string]string)
	for k, v := range data {
		dataStr[k] = fmt.Sprintf("%v", v)
	}

	message := &messaging.Message{
		Token: deviceToken,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: dataStr,
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
		APNS: &messaging.APNSConfig{
			Headers: map[string]string{
				"apns-priority": "10",
			},
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{
						Title: title,
						Body:  body,
					},
					Sound: "default",
				},
			},
		},
	}

	response, err := p.client.Send(ctx, message)
	if err != nil {
		// Check if it's a token error (should not retry)
		shouldRetry := !isTokenError(err)

		return &models.DeliveryResult{
			Success:      false,
			ErrorMessage: stringPtr(err.Error()),
			ErrorCode:    stringPtr("FCM_ERROR"),
			DeliveredAt:  time.Now(),
			ShouldRetry:  shouldRetry,
		}, err
	}

	return &models.DeliveryResult{
		Success:           true,
		ExternalMessageID: &response,
		DeliveredAt:       time.Now(),
		ShouldRetry:       false,
	}, nil
}

// GetChannel returns the channel this provider supports
func (p *FCMProvider) GetChannel() models.NotificationChannel {
	return models.ChannelPush
}

// IsHealthy checks if the FCM service is healthy
func (p *FCMProvider) IsHealthy(ctx context.Context) error {
	if p.client == nil {
		return fmt.Errorf("FCM client not initialized")
	}
	return nil
}

// isTokenError checks if the error is related to an invalid token
func isTokenError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return contains(errMsg, "invalid-registration-token") ||
		contains(errMsg, "registration-token-not-registered") ||
		contains(errMsg, "invalid-argument")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
