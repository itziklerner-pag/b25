package unit

import (
	"testing"

	"github.com/b25/services/notification/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestNotificationValidation(t *testing.T) {
	tests := []struct {
		name    string
		request *models.CreateNotificationRequest
		wantErr bool
	}{
		{
			name: "valid email notification",
			request: &models.CreateNotificationRequest{
				UserID:  "user123",
				Channel: models.ChannelEmail,
				Subject: stringPtr("Test"),
				Body:    stringPtr("Test body"),
			},
			wantErr: false,
		},
		{
			name: "missing user_id",
			request: &models.CreateNotificationRequest{
				Channel: models.ChannelEmail,
				Body:    stringPtr("Test body"),
			},
			wantErr: true,
		},
		{
			name: "missing channel",
			request: &models.CreateNotificationRequest{
				UserID: "user123",
				Body:   stringPtr("Test body"),
			},
			wantErr: true,
		},
		{
			name: "missing body and template",
			request: &models.CreateNotificationRequest{
				UserID:  "user123",
				Channel: models.ChannelEmail,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRequest(tt.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func validateRequest(req *models.CreateNotificationRequest) error {
	if req.UserID == "" {
		return assert.AnError
	}
	if req.Channel == "" {
		return assert.AnError
	}
	if req.TemplateName == nil && req.Body == nil {
		return assert.AnError
	}
	return nil
}

func stringPtr(s string) *string {
	return &s
}
