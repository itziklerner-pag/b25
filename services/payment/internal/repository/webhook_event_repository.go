package repository

import (
	"database/sql"
	"fmt"

	"github.com/yourorg/b25/services/payment/internal/database"
	"github.com/yourorg/b25/services/payment/internal/models"
)

type WebhookEventRepository struct {
	db *database.DB
}

func NewWebhookEventRepository(db *database.DB) *WebhookEventRepository {
	return &WebhookEventRepository{db: db}
}

func (r *WebhookEventRepository) Create(event *models.WebhookEvent) error {
	query := `
		INSERT INTO webhook_events (
			id, stripe_event_id, type, api_version, data, processed
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(
		query,
		event.ID, event.StripeEventID, event.Type,
		event.APIVersion, event.Data, event.Processed,
	)

	return err
}

func (r *WebhookEventRepository) GetByID(id string) (*models.WebhookEvent, error) {
	query := `
		SELECT id, stripe_event_id, type, api_version, data,
		       processed, processed_at, error, created_at
		FROM webhook_events
		WHERE id = $1
	`

	event := &models.WebhookEvent{}

	err := r.db.QueryRow(query, id).Scan(
		&event.ID, &event.StripeEventID, &event.Type,
		&event.APIVersion, &event.Data, &event.Processed,
		&event.ProcessedAt, &event.Error, &event.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("webhook event not found")
		}
		return nil, err
	}

	return event, nil
}

func (r *WebhookEventRepository) GetByStripeEventID(stripeEventID string) (*models.WebhookEvent, error) {
	query := `
		SELECT id, stripe_event_id, type, api_version, data,
		       processed, processed_at, error, created_at
		FROM webhook_events
		WHERE stripe_event_id = $1
	`

	event := &models.WebhookEvent{}

	err := r.db.QueryRow(query, stripeEventID).Scan(
		&event.ID, &event.StripeEventID, &event.Type,
		&event.APIVersion, &event.Data, &event.Processed,
		&event.ProcessedAt, &event.Error, &event.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Event doesn't exist yet, which is normal
		}
		return nil, err
	}

	return event, nil
}

func (r *WebhookEventRepository) MarkAsProcessed(id string) error {
	query := `
		UPDATE webhook_events
		SET processed = TRUE, processed_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *WebhookEventRepository) MarkAsFailedWithError(id, errorMsg string) error {
	query := `
		UPDATE webhook_events
		SET processed = TRUE, processed_at = NOW(), error = $1
		WHERE id = $2
	`
	_, err := r.db.Exec(query, errorMsg, id)
	return err
}

func (r *WebhookEventRepository) GetUnprocessed(limit int) ([]*models.WebhookEvent, error) {
	query := `
		SELECT id, stripe_event_id, type, api_version, data,
		       processed, processed_at, error, created_at
		FROM webhook_events
		WHERE processed = FALSE
		ORDER BY created_at ASC
		LIMIT $1
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*models.WebhookEvent

	for rows.Next() {
		event := &models.WebhookEvent{}

		err := rows.Scan(
			&event.ID, &event.StripeEventID, &event.Type,
			&event.APIVersion, &event.Data, &event.Processed,
			&event.ProcessedAt, &event.Error, &event.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	return events, nil
}
