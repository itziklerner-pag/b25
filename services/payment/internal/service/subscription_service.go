package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v76"
	"github.com/yourorg/b25/services/payment/internal/logger"
	"github.com/yourorg/b25/services/payment/internal/models"
	"github.com/yourorg/b25/services/payment/internal/payment"
	"github.com/yourorg/b25/services/payment/internal/repository"
)

type SubscriptionService struct {
	subRepo      *repository.SubscriptionRepository
	invRepo      *repository.InvoiceRepository
	stripeClient *payment.StripeClient
	logger       *logger.Logger
}

func NewSubscriptionService(
	subRepo *repository.SubscriptionRepository,
	invRepo *repository.InvoiceRepository,
	stripeClient *payment.StripeClient,
	logger *logger.Logger,
) *SubscriptionService {
	return &SubscriptionService{
		subRepo:      subRepo,
		invRepo:      invRepo,
		stripeClient: stripeClient,
		logger:       logger,
	}
}

// CreateSubscription creates a new subscription
func (s *SubscriptionService) CreateSubscription(req *models.CreateSubscriptionRequest) (*models.Subscription, error) {
	s.logger.Info("Creating subscription", "user_id", req.UserID, "price_id", req.PriceID)

	// Create subscription in Stripe
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(req.UserID), // Assumes customer ID = user ID or needs mapping
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(req.PriceID),
			},
		},
		DefaultPaymentMethod: stripe.String(req.PaymentMethod),
		PaymentBehavior:      stripe.String("default_incomplete"),
	}

	// Add trial period if specified
	if req.TrialDays > 0 {
		trialEnd := time.Now().AddDate(0, 0, req.TrialDays).Unix()
		params.TrialEnd = stripe.Int64(trialEnd)
	}

	// Add metadata
	if req.Metadata != nil {
		for k, v := range req.Metadata {
			params.AddMetadata(k, fmt.Sprintf("%v", v))
		}
	}
	params.AddMetadata("user_id", req.UserID)

	stripeSub, err := s.stripeClient.CreateSubscription(params)
	if err != nil {
		s.logger.Error("Failed to create subscription in Stripe", "error", err)
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	// Get price and product details
	price := stripeSub.Items.Data[0].Price
	product := price.Product

	// Store subscription in database
	metadataJSON, _ := json.Marshal(req.Metadata)
	subscription := &models.Subscription{
		ID:                   uuid.New().String(),
		UserID:               req.UserID,
		StripeSubscriptionID: stripeSub.ID,
		StripePriceID:        price.ID,
		StripeProductID:      product.ID,
		Status:               string(stripeSub.Status),
		PlanName:             product.Name,
		Amount:               price.UnitAmount,
		Currency:             string(price.Currency),
		Interval:             string(price.Recurring.Interval),
		IntervalCount:        int(price.Recurring.IntervalCount),
		CurrentPeriodStart:   time.Unix(stripeSub.CurrentPeriodStart, 0),
		CurrentPeriodEnd:     time.Unix(stripeSub.CurrentPeriodEnd, 0),
		CancelAtPeriodEnd:    stripeSub.CancelAtPeriodEnd,
		Metadata:             string(metadataJSON),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	if stripeSub.TrialEnd > 0 {
		trialEnd := time.Unix(stripeSub.TrialEnd, 0)
		subscription.TrialEnd = &trialEnd
	}

	if err := s.subRepo.Create(subscription); err != nil {
		s.logger.Error("Failed to save subscription", "error", err)
		return nil, fmt.Errorf("failed to save subscription: %w", err)
	}

	s.logger.Info("Subscription created successfully", "subscription_id", subscription.ID)
	return subscription, nil
}

// GetSubscription retrieves a subscription by ID
func (s *SubscriptionService) GetSubscription(id string) (*models.Subscription, error) {
	return s.subRepo.GetByID(id)
}

// GetUserSubscriptions retrieves all subscriptions for a user
func (s *SubscriptionService) GetUserSubscriptions(userID string) ([]*models.Subscription, error) {
	return s.subRepo.GetByUserID(userID)
}

// GetActiveSubscription retrieves the active subscription for a user
func (s *SubscriptionService) GetActiveSubscription(userID string) (*models.Subscription, error) {
	return s.subRepo.GetActiveByUserID(userID)
}

// CancelSubscription cancels a subscription
func (s *SubscriptionService) CancelSubscription(id string, immediate bool) (*models.Subscription, error) {
	s.logger.Info("Canceling subscription", "id", id, "immediate", immediate)

	// Get subscription from database
	subscription, err := s.subRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Cancel in Stripe
	if immediate {
		_, err = s.stripeClient.CancelSubscription(subscription.StripeSubscriptionID)
	} else {
		// Cancel at period end
		params := &stripe.SubscriptionParams{
			CancelAtPeriodEnd: stripe.Bool(true),
		}
		_, err = s.stripeClient.UpdateSubscription(subscription.StripeSubscriptionID, params)
	}

	if err != nil {
		s.logger.Error("Failed to cancel subscription in Stripe", "error", err)
		return nil, fmt.Errorf("failed to cancel subscription: %w", err)
	}

	// Update in database
	if immediate {
		now := time.Now()
		subscription.Status = models.SubscriptionStatusCanceled
		subscription.CanceledAt = &now
		subscription.EndedAt = &now
	} else {
		now := time.Now()
		subscription.CancelAtPeriodEnd = true
		subscription.CanceledAt = &now
	}

	if err := s.subRepo.Update(subscription); err != nil {
		s.logger.Error("Failed to update subscription", "error", err)
		return nil, err
	}

	s.logger.Info("Subscription canceled successfully", "id", id)
	return subscription, nil
}

// UpdateSubscriptionFromWebhook updates a subscription from a webhook event
func (s *SubscriptionService) UpdateSubscriptionFromWebhook(stripeSub *stripe.Subscription) error {
	s.logger.Info("Updating subscription from webhook", "stripe_subscription_id", stripeSub.ID)

	subscription, err := s.subRepo.GetByStripeSubscriptionID(stripeSub.ID)
	if err != nil {
		s.logger.Error("Subscription not found", "error", err)
		return err
	}

	subscription.Status = string(stripeSub.Status)
	subscription.CurrentPeriodStart = time.Unix(stripeSub.CurrentPeriodStart, 0)
	subscription.CurrentPeriodEnd = time.Unix(stripeSub.CurrentPeriodEnd, 0)
	subscription.CancelAtPeriodEnd = stripeSub.CancelAtPeriodEnd

	if stripeSub.CanceledAt > 0 {
		canceledAt := time.Unix(stripeSub.CanceledAt, 0)
		subscription.CanceledAt = &canceledAt
	}

	if stripeSub.EndedAt > 0 {
		endedAt := time.Unix(stripeSub.EndedAt, 0)
		subscription.EndedAt = &endedAt
	}

	return s.subRepo.Update(subscription)
}
