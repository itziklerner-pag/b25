package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/yourorg/b25/services/payment/internal/database"
	"github.com/yourorg/b25/services/payment/internal/models"
)

type SubscriptionRepository struct {
	db *database.DB
}

func NewSubscriptionRepository(db *database.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Create(sub *models.Subscription) error {
	metadataJSON, _ := json.Marshal(sub.Metadata)

	query := `
		INSERT INTO subscriptions (
			id, user_id, stripe_subscription_id, stripe_price_id,
			stripe_product_id, status, plan_name, amount, currency,
			interval, interval_count, trial_end, current_period_start,
			current_period_end, cancel_at_period_end, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	_, err := r.db.Exec(
		query,
		sub.ID, sub.UserID, sub.StripeSubscriptionID, sub.StripePriceID,
		sub.StripeProductID, sub.Status, sub.PlanName, sub.Amount, sub.Currency,
		sub.Interval, sub.IntervalCount, sub.TrialEnd, sub.CurrentPeriodStart,
		sub.CurrentPeriodEnd, sub.CancelAtPeriodEnd, metadataJSON,
	)

	return err
}

func (r *SubscriptionRepository) GetByID(id string) (*models.Subscription, error) {
	query := `
		SELECT id, user_id, stripe_subscription_id, stripe_price_id,
		       stripe_product_id, status, plan_name, amount, currency,
		       interval, interval_count, trial_end, current_period_start,
		       current_period_end, cancel_at_period_end, canceled_at,
		       ended_at, metadata, created_at, updated_at
		FROM subscriptions
		WHERE id = $1
	`

	sub := &models.Subscription{}
	var metadata sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&sub.ID, &sub.UserID, &sub.StripeSubscriptionID, &sub.StripePriceID,
		&sub.StripeProductID, &sub.Status, &sub.PlanName, &sub.Amount,
		&sub.Currency, &sub.Interval, &sub.IntervalCount, &sub.TrialEnd,
		&sub.CurrentPeriodStart, &sub.CurrentPeriodEnd, &sub.CancelAtPeriodEnd,
		&sub.CanceledAt, &sub.EndedAt, &metadata, &sub.CreatedAt, &sub.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("subscription not found")
		}
		return nil, err
	}

	if metadata.Valid {
		sub.Metadata = metadata.String
	}

	return sub, nil
}

func (r *SubscriptionRepository) GetByStripeSubscriptionID(stripeSubID string) (*models.Subscription, error) {
	query := `
		SELECT id, user_id, stripe_subscription_id, stripe_price_id,
		       stripe_product_id, status, plan_name, amount, currency,
		       interval, interval_count, trial_end, current_period_start,
		       current_period_end, cancel_at_period_end, canceled_at,
		       ended_at, metadata, created_at, updated_at
		FROM subscriptions
		WHERE stripe_subscription_id = $1
	`

	sub := &models.Subscription{}
	var metadata sql.NullString

	err := r.db.QueryRow(query, stripeSubID).Scan(
		&sub.ID, &sub.UserID, &sub.StripeSubscriptionID, &sub.StripePriceID,
		&sub.StripeProductID, &sub.Status, &sub.PlanName, &sub.Amount,
		&sub.Currency, &sub.Interval, &sub.IntervalCount, &sub.TrialEnd,
		&sub.CurrentPeriodStart, &sub.CurrentPeriodEnd, &sub.CancelAtPeriodEnd,
		&sub.CanceledAt, &sub.EndedAt, &metadata, &sub.CreatedAt, &sub.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("subscription not found")
		}
		return nil, err
	}

	if metadata.Valid {
		sub.Metadata = metadata.String
	}

	return sub, nil
}

func (r *SubscriptionRepository) GetByUserID(userID string) ([]*models.Subscription, error) {
	query := `
		SELECT id, user_id, stripe_subscription_id, stripe_price_id,
		       stripe_product_id, status, plan_name, amount, currency,
		       interval, interval_count, trial_end, current_period_start,
		       current_period_end, cancel_at_period_end, canceled_at,
		       ended_at, metadata, created_at, updated_at
		FROM subscriptions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []*models.Subscription

	for rows.Next() {
		sub := &models.Subscription{}
		var metadata sql.NullString

		err := rows.Scan(
			&sub.ID, &sub.UserID, &sub.StripeSubscriptionID, &sub.StripePriceID,
			&sub.StripeProductID, &sub.Status, &sub.PlanName, &sub.Amount,
			&sub.Currency, &sub.Interval, &sub.IntervalCount, &sub.TrialEnd,
			&sub.CurrentPeriodStart, &sub.CurrentPeriodEnd, &sub.CancelAtPeriodEnd,
			&sub.CanceledAt, &sub.EndedAt, &metadata, &sub.CreatedAt, &sub.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if metadata.Valid {
			sub.Metadata = metadata.String
		}

		subscriptions = append(subscriptions, sub)
	}

	return subscriptions, nil
}

func (r *SubscriptionRepository) GetActiveByUserID(userID string) (*models.Subscription, error) {
	query := `
		SELECT id, user_id, stripe_subscription_id, stripe_price_id,
		       stripe_product_id, status, plan_name, amount, currency,
		       interval, interval_count, trial_end, current_period_start,
		       current_period_end, cancel_at_period_end, canceled_at,
		       ended_at, metadata, created_at, updated_at
		FROM subscriptions
		WHERE user_id = $1 AND status = 'active'
		ORDER BY created_at DESC
		LIMIT 1
	`

	sub := &models.Subscription{}
	var metadata sql.NullString

	err := r.db.QueryRow(query, userID).Scan(
		&sub.ID, &sub.UserID, &sub.StripeSubscriptionID, &sub.StripePriceID,
		&sub.StripeProductID, &sub.Status, &sub.PlanName, &sub.Amount,
		&sub.Currency, &sub.Interval, &sub.IntervalCount, &sub.TrialEnd,
		&sub.CurrentPeriodStart, &sub.CurrentPeriodEnd, &sub.CancelAtPeriodEnd,
		&sub.CanceledAt, &sub.EndedAt, &metadata, &sub.CreatedAt, &sub.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no active subscription found")
		}
		return nil, err
	}

	if metadata.Valid {
		sub.Metadata = metadata.String
	}

	return sub, nil
}

func (r *SubscriptionRepository) Update(sub *models.Subscription) error {
	metadataJSON, _ := json.Marshal(sub.Metadata)

	query := `
		UPDATE subscriptions
		SET status = $1, current_period_start = $2, current_period_end = $3,
		    cancel_at_period_end = $4, canceled_at = $5, ended_at = $6,
		    metadata = $7
		WHERE id = $8
	`

	_, err := r.db.Exec(
		query,
		sub.Status, sub.CurrentPeriodStart, sub.CurrentPeriodEnd,
		sub.CancelAtPeriodEnd, sub.CanceledAt, sub.EndedAt,
		metadataJSON, sub.ID,
	)

	return err
}

func (r *SubscriptionRepository) UpdateStatus(id, status string) error {
	query := `UPDATE subscriptions SET status = $1 WHERE id = $2`
	_, err := r.db.Exec(query, status, id)
	return err
}

func (r *SubscriptionRepository) Cancel(id string) error {
	query := `
		UPDATE subscriptions
		SET status = 'canceled', cancel_at_period_end = TRUE, canceled_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(query, id)
	return err
}
