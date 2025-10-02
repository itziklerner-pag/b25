package providers

import (
	"context"
	"fmt"
	"time"

	"github.com/b25/services/notification/internal/config"
	"github.com/b25/services/notification/internal/models"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// SendGridProvider implements email delivery using SendGrid
type SendGridProvider struct {
	client      *sendgrid.Client
	fromAddress string
	fromName    string
}

// NewSendGridProvider creates a new SendGrid provider
func NewSendGridProvider(cfg *config.EmailConfig) (EmailProvider, error) {
	if cfg.SendGrid.APIKey == "" {
		return nil, fmt.Errorf("sendgrid api key is required")
	}

	return &SendGridProvider{
		client:      sendgrid.NewSendClient(cfg.SendGrid.APIKey),
		fromAddress: cfg.FromAddress,
		fromName:    cfg.FromName,
	}, nil
}

// Send sends an email notification
func (p *SendGridProvider) Send(ctx context.Context, notification *models.Notification) (*models.DeliveryResult, error) {
	if notification.RecipientAddress == nil {
		return nil, fmt.Errorf("recipient address is required for email")
	}

	subject := ""
	if notification.Subject != nil {
		subject = *notification.Subject
	}

	// Assume HTML if body contains HTML tags
	isHTML := len(notification.Body) > 0 && (notification.Body[0] == '<')

	return p.SendEmail(ctx, *notification.RecipientAddress, subject, notification.Body, isHTML)
}

// SendEmail sends an email using SendGrid
func (p *SendGridProvider) SendEmail(ctx context.Context, to, subject, body string, isHTML bool) (*models.DeliveryResult, error) {
	from := mail.NewEmail(p.fromName, p.fromAddress)
	toEmail := mail.NewEmail("", to)

	var message *mail.SGMailV3
	if isHTML {
		message = mail.NewSingleEmail(from, subject, toEmail, "", body)
	} else {
		message = mail.NewSingleEmail(from, subject, toEmail, body, "")
	}

	response, err := p.client.SendWithContext(ctx, message)
	if err != nil {
		return &models.DeliveryResult{
			Success:      false,
			ErrorMessage: stringPtr(err.Error()),
			ErrorCode:    stringPtr("SENDGRID_ERROR"),
			DeliveredAt:  time.Now(),
			ShouldRetry:  true,
		}, err
	}

	// SendGrid returns 202 for accepted messages
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		// Extract message ID from headers
		messageID := response.Headers.Get("X-Message-Id")

		return &models.DeliveryResult{
			Success:           true,
			ExternalMessageID: &messageID,
			DeliveredAt:       time.Now(),
			ShouldRetry:       false,
		}, nil
	}

	// Handle error responses
	shouldRetry := response.StatusCode >= 500 || response.StatusCode == 429

	return &models.DeliveryResult{
		Success:      false,
		ErrorMessage: stringPtr(response.Body),
		ErrorCode:    stringPtr(fmt.Sprintf("HTTP_%d", response.StatusCode)),
		DeliveredAt:  time.Now(),
		ShouldRetry:  shouldRetry,
	}, fmt.Errorf("sendgrid returned status %d: %s", response.StatusCode, response.Body)
}

// GetChannel returns the channel this provider supports
func (p *SendGridProvider) GetChannel() models.NotificationChannel {
	return models.ChannelEmail
}

// IsHealthy checks if the SendGrid service is healthy
func (p *SendGridProvider) IsHealthy(ctx context.Context) error {
	// SendGrid doesn't have a health check endpoint
	// We can just verify the API key is set
	if p.client == nil {
		return fmt.Errorf("sendgrid client not initialized")
	}
	return nil
}

func stringPtr(s string) *string {
	return &s
}
