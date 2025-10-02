package providers

import (
	"context"
	"fmt"
	"time"

	"github.com/b25/services/notification/internal/config"
	"github.com/b25/services/notification/internal/models"
	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

// TwilioProvider implements SMS delivery using Twilio
type TwilioProvider struct {
	client     *twilio.RestClient
	fromNumber string
}

// NewTwilioProvider creates a new Twilio provider
func NewTwilioProvider(cfg *config.SMSConfig) (SMSProvider, error) {
	if cfg.Twilio.AccountSID == "" || cfg.Twilio.AuthToken == "" {
		return nil, fmt.Errorf("twilio credentials are required")
	}

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: cfg.Twilio.AccountSID,
		Password: cfg.Twilio.AuthToken,
	})

	return &TwilioProvider{
		client:     client,
		fromNumber: cfg.Twilio.FromNumber,
	}, nil
}

// Send sends an SMS notification
func (p *TwilioProvider) Send(ctx context.Context, notification *models.Notification) (*models.DeliveryResult, error) {
	if notification.RecipientAddress == nil {
		return nil, fmt.Errorf("recipient phone number is required for SMS")
	}

	return p.SendSMS(ctx, *notification.RecipientAddress, notification.Body)
}

// SendSMS sends an SMS using Twilio
func (p *TwilioProvider) SendSMS(ctx context.Context, to, message string) (*models.DeliveryResult, error) {
	params := &twilioApi.CreateMessageParams{}
	params.SetTo(to)
	params.SetFrom(p.fromNumber)
	params.SetBody(message)

	resp, err := p.client.Api.CreateMessage(params)
	if err != nil {
		return &models.DeliveryResult{
			Success:      false,
			ErrorMessage: stringPtr(err.Error()),
			ErrorCode:    stringPtr("TWILIO_ERROR"),
			DeliveredAt:  time.Now(),
			ShouldRetry:  true,
		}, err
	}

	// Check if message was accepted
	if resp.Status == nil {
		return &models.DeliveryResult{
			Success:      false,
			ErrorMessage: stringPtr("no status returned from Twilio"),
			ErrorCode:    stringPtr("TWILIO_NO_STATUS"),
			DeliveredAt:  time.Now(),
			ShouldRetry:  true,
		}, fmt.Errorf("no status returned from Twilio")
	}

	status := *resp.Status
	success := status == "queued" || status == "sent" || status == "delivered"

	var messageID string
	if resp.Sid != nil {
		messageID = *resp.Sid
	}

	result := &models.DeliveryResult{
		Success:           success,
		ExternalMessageID: &messageID,
		DeliveredAt:       time.Now(),
		ShouldRetry:       !success && (status == "failed" || status == "undelivered"),
	}

	if !success {
		errorMsg := fmt.Sprintf("Twilio status: %s", status)
		if resp.ErrorMessage != nil {
			errorMsg = *resp.ErrorMessage
		}
		result.ErrorMessage = &errorMsg

		if resp.ErrorCode != nil {
			errorCode := fmt.Sprintf("TWILIO_%d", *resp.ErrorCode)
			result.ErrorCode = &errorCode
		}

		return result, fmt.Errorf("SMS failed with status: %s", status)
	}

	return result, nil
}

// GetChannel returns the channel this provider supports
func (p *TwilioProvider) GetChannel() models.NotificationChannel {
	return models.ChannelSMS
}

// IsHealthy checks if the Twilio service is healthy
func (p *TwilioProvider) IsHealthy(ctx context.Context) error {
	if p.client == nil {
		return fmt.Errorf("twilio client not initialized")
	}
	// Could make a test API call here, but for now just check client exists
	return nil
}
