package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stripe/stripe-go/v76"
	"github.com/yourorg/b25/services/payment/internal/logger"
	"github.com/yourorg/b25/services/payment/internal/models"
	"github.com/yourorg/b25/services/payment/internal/payment"
	"github.com/yourorg/b25/services/payment/internal/repository"
)

type PaymentService struct {
	txRepo        *repository.TransactionRepository
	pmRepo        *repository.PaymentMethodRepository
	stripeClient  *payment.StripeClient
	redisClient   *redis.Client
	logger        *logger.Logger
}

func NewPaymentService(
	txRepo *repository.TransactionRepository,
	pmRepo *repository.PaymentMethodRepository,
	stripeClient *payment.StripeClient,
	redisClient *redis.Client,
	logger *logger.Logger,
) *PaymentService {
	return &PaymentService{
		txRepo:       txRepo,
		pmRepo:       pmRepo,
		stripeClient: stripeClient,
		redisClient:  redisClient,
		logger:       logger,
	}
}

// CreatePayment creates a new payment
func (s *PaymentService) CreatePayment(req *models.CreatePaymentRequest) (*models.Transaction, error) {
	s.logger.Info("Creating payment", "user_id", req.UserID, "amount", req.Amount)

	// Create payment intent in Stripe
	params := &stripe.PaymentIntentParams{
		Amount:      stripe.Int64(req.Amount),
		Currency:    stripe.String(req.Currency),
		Description: stripe.String(req.Description),
	}

	if req.PaymentMethod != "" {
		params.PaymentMethod = stripe.String(req.PaymentMethod)
		params.Confirm = stripe.Bool(true)
	}

	// Add metadata
	if req.Metadata != nil {
		for k, v := range req.Metadata {
			params.AddMetadata(k, fmt.Sprintf("%v", v))
		}
	}
	params.AddMetadata("user_id", req.UserID)

	pi, err := s.stripeClient.CreatePaymentIntent(params)
	if err != nil {
		s.logger.Error("Failed to create payment intent", "error", err)
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Store transaction in database
	metadataJSON, _ := json.Marshal(req.Metadata)
	tx := &models.Transaction{
		ID:              uuid.New().String(),
		UserID:          req.UserID,
		StripePaymentID: pi.ID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		Status:          string(pi.Status),
		PaymentMethod:   req.PaymentMethod,
		Description:     req.Description,
		Metadata:        string(metadataJSON),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.txRepo.Create(tx); err != nil {
		s.logger.Error("Failed to save transaction", "error", err)
		// Payment intent was created in Stripe, but we failed to save it
		// We should cancel it to avoid orphaned payments
		s.stripeClient.CancelPaymentIntent(pi.ID)
		return nil, fmt.Errorf("failed to save transaction: %w", err)
	}

	s.logger.Info("Payment created successfully", "transaction_id", tx.ID)
	return tx, nil
}

// GetTransaction retrieves a transaction by ID
func (s *PaymentService) GetTransaction(id string) (*models.Transaction, error) {
	// Try cache first
	ctx := context.Background()
	cacheKey := fmt.Sprintf("transaction:%s", id)

	cachedData, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var tx models.Transaction
		if err := json.Unmarshal([]byte(cachedData), &tx); err == nil {
			return &tx, nil
		}
	}

	// Get from database
	tx, err := s.txRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Cache the result
	txJSON, _ := json.Marshal(tx)
	s.redisClient.Set(ctx, cacheKey, txJSON, 5*time.Minute)

	return tx, nil
}

// GetUserTransactions retrieves all transactions for a user
func (s *PaymentService) GetUserTransactions(userID string, limit, offset int) ([]*models.Transaction, error) {
	return s.txRepo.GetByUserID(userID, limit, offset)
}

// UpdateTransactionStatus updates the status of a transaction
func (s *PaymentService) UpdateTransactionStatus(id, status string) error {
	s.logger.Info("Updating transaction status", "id", id, "status", status)

	err := s.txRepo.UpdateStatus(id, status)
	if err != nil {
		s.logger.Error("Failed to update transaction status", "error", err)
		return err
	}

	// Invalidate cache
	ctx := context.Background()
	cacheKey := fmt.Sprintf("transaction:%s", id)
	s.redisClient.Del(ctx, cacheKey)

	return nil
}

// AttachPaymentMethod attaches a payment method to a user
func (s *PaymentService) AttachPaymentMethod(req *models.PaymentMethodRequest) (*models.PaymentMethod, error) {
	s.logger.Info("Attaching payment method", "user_id", req.UserID)

	// Get or create Stripe customer for user
	customerID, err := s.getOrCreateStripeCustomer(req.UserID)
	if err != nil {
		return nil, err
	}

	// Attach payment method to customer in Stripe
	pm, err := s.stripeClient.AttachPaymentMethod(req.PaymentMethod, customerID)
	if err != nil {
		s.logger.Error("Failed to attach payment method", "error", err)
		return nil, fmt.Errorf("failed to attach payment method: %w", err)
	}

	// Store payment method in database
	paymentMethod := &models.PaymentMethod{
		ID:                    uuid.New().String(),
		UserID:                req.UserID,
		StripePaymentMethodID: pm.ID,
		Type:                  string(pm.Type),
		IsDefault:             req.SetAsDefault,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	// Extract card details if applicable
	if pm.Card != nil {
		brand := string(pm.Card.Brand)
		last4 := pm.Card.Last4
		expMonth := int(pm.Card.ExpMonth)
		expYear := int(pm.Card.ExpYear)

		paymentMethod.CardBrand = &brand
		paymentMethod.CardLast4 = &last4
		paymentMethod.CardExpMonth = &expMonth
		paymentMethod.CardExpYear = &expYear
	}

	if err := s.pmRepo.Create(paymentMethod); err != nil {
		s.logger.Error("Failed to save payment method", "error", err)
		return nil, fmt.Errorf("failed to save payment method: %w", err)
	}

	// Set as default if requested
	if req.SetAsDefault {
		if err := s.pmRepo.SetAsDefault(paymentMethod.ID, req.UserID); err != nil {
			s.logger.Error("Failed to set payment method as default", "error", err)
		}
	}

	s.logger.Info("Payment method attached successfully", "id", paymentMethod.ID)
	return paymentMethod, nil
}

// GetUserPaymentMethods retrieves all payment methods for a user
func (s *PaymentService) GetUserPaymentMethods(userID string) ([]*models.PaymentMethod, error) {
	return s.pmRepo.GetByUserID(userID)
}

// DetachPaymentMethod removes a payment method
func (s *PaymentService) DetachPaymentMethod(id, userID string) error {
	s.logger.Info("Detaching payment method", "id", id)

	pm, err := s.pmRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Verify ownership
	if pm.UserID != userID {
		return fmt.Errorf("unauthorized: payment method does not belong to user")
	}

	// Detach from Stripe
	if _, err := s.stripeClient.DetachPaymentMethod(pm.StripePaymentMethodID); err != nil {
		s.logger.Error("Failed to detach payment method from Stripe", "error", err)
		return err
	}

	// Delete from database
	if err := s.pmRepo.Delete(id); err != nil {
		s.logger.Error("Failed to delete payment method", "error", err)
		return err
	}

	s.logger.Info("Payment method detached successfully", "id", id)
	return nil
}

// getOrCreateStripeCustomer gets or creates a Stripe customer for a user
func (s *PaymentService) getOrCreateStripeCustomer(userID string) (string, error) {
	// Check cache for customer ID
	ctx := context.Background()
	cacheKey := fmt.Sprintf("stripe_customer:%s", userID)

	customerID, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil && customerID != "" {
		return customerID, nil
	}

	// Create new customer
	params := &stripe.CustomerParams{}
	params.AddMetadata("user_id", userID)

	customer, err := s.stripeClient.CreateCustomer(params)
	if err != nil {
		return "", fmt.Errorf("failed to create Stripe customer: %w", err)
	}

	// Cache customer ID
	s.redisClient.Set(ctx, cacheKey, customer.ID, 24*time.Hour)

	return customer.ID, nil
}
