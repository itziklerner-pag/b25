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

type RefundService struct {
	txRepo       *repository.TransactionRepository
	stripeClient *payment.StripeClient
	logger       *logger.Logger
}

func NewRefundService(
	txRepo *repository.TransactionRepository,
	stripeClient *payment.StripeClient,
	logger *logger.Logger,
) *RefundService {
	return &RefundService{
		txRepo:       txRepo,
		stripeClient: stripeClient,
		logger:       logger,
	}
}

// CreateRefund creates a new refund
func (s *RefundService) CreateRefund(req *models.CreateRefundRequest) (*models.Refund, error) {
	s.logger.Info("Creating refund", "transaction_id", req.TransactionID)

	// Get original transaction
	tx, err := s.txRepo.GetByID(req.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("transaction not found: %w", err)
	}

	// Check if transaction can be refunded
	if tx.Status != models.TransactionStatusSucceeded {
		return nil, fmt.Errorf("transaction cannot be refunded: status is %s", tx.Status)
	}

	// Determine refund amount
	refundAmount := req.Amount
	if refundAmount == 0 || refundAmount > tx.Amount {
		refundAmount = tx.Amount // Full refund
	}

	// Check if already fully refunded
	if tx.RefundedAmount >= tx.Amount {
		return nil, fmt.Errorf("transaction is already fully refunded")
	}

	// Check if refund amount exceeds remaining amount
	remainingAmount := tx.Amount - tx.RefundedAmount
	if refundAmount > remainingAmount {
		return nil, fmt.Errorf("refund amount exceeds remaining refundable amount")
	}

	// Create refund in Stripe
	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(tx.StripePaymentID),
		Amount:        stripe.Int64(refundAmount),
		Reason:        stripe.String(req.Reason),
	}

	stripeRefund, err := s.stripeClient.CreateRefund(params)
	if err != nil {
		s.logger.Error("Failed to create refund in Stripe", "error", err)
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}

	// Create refund record
	metadataJSON, _ := json.Marshal(map[string]interface{}{
		"transaction_id": req.TransactionID,
		"reason":         req.Reason,
	})

	refund := &models.Refund{
		ID:             uuid.New().String(),
		TransactionID:  req.TransactionID,
		StripeRefundID: stripeRefund.ID,
		Amount:         refundAmount,
		Currency:       tx.Currency,
		Reason:         req.Reason,
		Status:         string(stripeRefund.Status),
		Metadata:       string(metadataJSON),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Update transaction with refunded amount
	newRefundedAmount := tx.RefundedAmount + refundAmount
	if err := s.txRepo.MarkAsRefunded(tx.ID, newRefundedAmount); err != nil {
		s.logger.Error("Failed to update transaction refund amount", "error", err)
		return nil, err
	}

	s.logger.Info("Refund created successfully", "refund_id", refund.ID)
	return refund, nil
}

// ProcessRefundWebhook processes a refund webhook event
func (s *RefundService) ProcessRefundWebhook(stripeRefund *stripe.Refund) error {
	s.logger.Info("Processing refund webhook", "stripe_refund_id", stripeRefund.ID)

	// Find the transaction
	tx, err := s.txRepo.GetByStripePaymentID(stripeRefund.PaymentIntent.ID)
	if err != nil {
		s.logger.Error("Transaction not found for refund", "error", err)
		return err
	}

	// Update transaction status based on refund status
	if stripeRefund.Status == "succeeded" {
		newRefundedAmount := tx.RefundedAmount + stripeRefund.Amount
		if err := s.txRepo.MarkAsRefunded(tx.ID, newRefundedAmount); err != nil {
			s.logger.Error("Failed to update transaction refund amount", "error", err)
			return err
		}
	}

	return nil
}
