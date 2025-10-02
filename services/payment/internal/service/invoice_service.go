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

type InvoiceService struct {
	invRepo      *repository.InvoiceRepository
	txRepo       *repository.TransactionRepository
	stripeClient *payment.StripeClient
	logger       *logger.Logger
}

func NewInvoiceService(
	invRepo *repository.InvoiceRepository,
	txRepo *repository.TransactionRepository,
	stripeClient *payment.StripeClient,
	logger *logger.Logger,
) *InvoiceService {
	return &InvoiceService{
		invRepo:      invRepo,
		txRepo:       txRepo,
		stripeClient: stripeClient,
		logger:       logger,
	}
}

// GetInvoice retrieves an invoice by ID
func (s *InvoiceService) GetInvoice(id string) (*models.Invoice, error) {
	return s.invRepo.GetByID(id)
}

// GetUserInvoices retrieves all invoices for a user
func (s *InvoiceService) GetUserInvoices(userID string, limit, offset int) ([]*models.Invoice, error) {
	return s.invRepo.GetByUserID(userID, limit, offset)
}

// GetSubscriptionInvoices retrieves all invoices for a subscription
func (s *InvoiceService) GetSubscriptionInvoices(subscriptionID string) ([]*models.Invoice, error) {
	return s.invRepo.GetBySubscriptionID(subscriptionID)
}

// CreateInvoiceFromStripe creates an invoice from a Stripe invoice
func (s *InvoiceService) CreateInvoiceFromStripe(stripeInv *stripe.Invoice) (*models.Invoice, error) {
	s.logger.Info("Creating invoice from Stripe", "stripe_invoice_id", stripeInv.ID)

	// Check if invoice already exists
	existingInv, err := s.invRepo.GetByStripeInvoiceID(stripeInv.ID)
	if err == nil && existingInv != nil {
		s.logger.Info("Invoice already exists", "id", existingInv.ID)
		return existingInv, nil
	}

	userID := ""
	if stripeInv.Customer != nil {
		userID = stripeInv.Customer.ID
	}

	var subscriptionID *string
	if stripeInv.Subscription != nil {
		subID := stripeInv.Subscription.ID
		subscriptionID = &subID
	}

	invoice := &models.Invoice{
		ID:               uuid.New().String(),
		UserID:           userID,
		StripeInvoiceID:  stripeInv.ID,
		SubscriptionID:   subscriptionID,
		Number:           stripeInv.Number,
		Status:           string(stripeInv.Status),
		AmountDue:        stripeInv.AmountDue,
		AmountPaid:       stripeInv.AmountPaid,
		AmountRemaining:  stripeInv.AmountRemaining,
		Currency:         string(stripeInv.Currency),
		Description:      stripeInv.Description,
		HostedInvoiceURL: stripeInv.HostedInvoiceURL,
		InvoicePDF:       stripeInv.InvoicePDF,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if stripeInv.DueDate > 0 {
		dueDate := time.Unix(stripeInv.DueDate, 0)
		invoice.DueDate = &dueDate
	}

	if stripeInv.StatusTransitions != nil && stripeInv.StatusTransitions.PaidAt > 0 {
		paidAt := time.Unix(stripeInv.StatusTransitions.PaidAt, 0)
		invoice.PaidAt = &paidAt
	}

	if stripeInv.StatusTransitions != nil && stripeInv.StatusTransitions.VoidedAt > 0 {
		voidedAt := time.Unix(stripeInv.StatusTransitions.VoidedAt, 0)
		invoice.VoidedAt = &voidedAt
	}

	if err := s.invRepo.Create(invoice); err != nil {
		s.logger.Error("Failed to create invoice", "error", err)
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	s.logger.Info("Invoice created successfully", "invoice_id", invoice.ID)
	return invoice, nil
}

// UpdateInvoiceFromWebhook updates an invoice from a webhook event
func (s *InvoiceService) UpdateInvoiceFromWebhook(stripeInv *stripe.Invoice) error {
	s.logger.Info("Updating invoice from webhook", "stripe_invoice_id", stripeInv.ID)

	invoice, err := s.invRepo.GetByStripeInvoiceID(stripeInv.ID)
	if err != nil {
		// Invoice doesn't exist, create it
		_, err = s.CreateInvoiceFromStripe(stripeInv)
		return err
	}

	invoice.Status = string(stripeInv.Status)
	invoice.AmountPaid = stripeInv.AmountPaid
	invoice.AmountRemaining = stripeInv.AmountRemaining

	if stripeInv.StatusTransitions != nil && stripeInv.StatusTransitions.PaidAt > 0 {
		paidAt := time.Unix(stripeInv.StatusTransitions.PaidAt, 0)
		invoice.PaidAt = &paidAt
	}

	if stripeInv.StatusTransitions != nil && stripeInv.StatusTransitions.VoidedAt > 0 {
		voidedAt := time.Unix(stripeInv.StatusTransitions.VoidedAt, 0)
		invoice.VoidedAt = &voidedAt
	}

	return s.invRepo.Update(invoice)
}

// MarkInvoiceAsPaid marks an invoice as paid
func (s *InvoiceService) MarkInvoiceAsPaid(invoiceID string, paymentIntentID string) error {
	s.logger.Info("Marking invoice as paid", "invoice_id", invoiceID)

	invoice, err := s.invRepo.GetByID(invoiceID)
	if err != nil {
		return err
	}

	// Create transaction record for the payment
	metadataJSON, _ := json.Marshal(map[string]interface{}{
		"invoice_id": invoiceID,
	})

	tx := &models.Transaction{
		ID:              uuid.New().String(),
		UserID:          invoice.UserID,
		StripePaymentID: paymentIntentID,
		Amount:          invoice.AmountDue,
		Currency:        invoice.Currency,
		Status:          models.TransactionStatusSucceeded,
		Description:     fmt.Sprintf("Payment for invoice %s", invoice.Number),
		Metadata:        string(metadataJSON),
		InvoiceID:       &invoice.ID,
		SubscriptionID:  invoice.SubscriptionID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.txRepo.Create(tx); err != nil {
		s.logger.Error("Failed to create transaction", "error", err)
		return err
	}

	// Mark invoice as paid
	if err := s.invRepo.MarkAsPaid(invoiceID); err != nil {
		s.logger.Error("Failed to mark invoice as paid", "error", err)
		return err
	}

	return nil
}
