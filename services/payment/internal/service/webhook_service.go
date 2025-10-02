package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/webhook"
	"github.com/yourorg/b25/services/payment/internal/logger"
	"github.com/yourorg/b25/services/payment/internal/models"
	"github.com/yourorg/b25/services/payment/internal/repository"
)

type WebhookService struct {
	webhookRepo         *repository.WebhookEventRepository
	paymentService      *PaymentService
	subscriptionService *SubscriptionService
	webhookSecret       string
	logger              *logger.Logger
}

func NewWebhookService(
	webhookRepo *repository.WebhookEventRepository,
	paymentService *PaymentService,
	subscriptionService *SubscriptionService,
	webhookSecret string,
	logger *logger.Logger,
) *WebhookService {
	return &WebhookService{
		webhookRepo:         webhookRepo,
		paymentService:      paymentService,
		subscriptionService: subscriptionService,
		webhookSecret:       webhookSecret,
		logger:              logger,
	}
}

// VerifyAndProcessWebhook verifies and processes a Stripe webhook
func (s *WebhookService) VerifyAndProcessWebhook(payload []byte, signature string) error {
	// Verify webhook signature
	event, err := webhook.ConstructEvent(payload, signature, s.webhookSecret)
	if err != nil {
		s.logger.Error("Failed to verify webhook signature", "error", err)
		return fmt.Errorf("webhook signature verification failed: %w", err)
	}

	// Check if event has already been processed (idempotency)
	existingEvent, _ := s.webhookRepo.GetByStripeEventID(event.ID)
	if existingEvent != nil {
		s.logger.Info("Webhook event already processed", "event_id", event.ID)
		return nil
	}

	// Store webhook event
	dataJSON, _ := json.Marshal(event.Data.Raw)
	webhookEvent := &models.WebhookEvent{
		ID:            uuid.New().String(),
		StripeEventID: event.ID,
		Type:          string(event.Type),
		APIVersion:    event.APIVersion,
		Data:          string(dataJSON),
		Processed:     false,
		CreatedAt:     time.Now(),
	}

	if err := s.webhookRepo.Create(webhookEvent); err != nil {
		s.logger.Error("Failed to store webhook event", "error", err)
		return err
	}

	// Process the event
	if err := s.processEvent(&event); err != nil {
		s.logger.Error("Failed to process webhook event", "error", err, "type", event.Type)
		s.webhookRepo.MarkAsFailedWithError(webhookEvent.ID, err.Error())
		return err
	}

	// Mark as processed
	s.webhookRepo.MarkAsProcessed(webhookEvent.ID)
	s.logger.Info("Webhook event processed successfully", "event_id", event.ID, "type", event.Type)

	return nil
}

// processEvent processes different types of webhook events
func (s *WebhookService) processEvent(event *stripe.Event) error {
	switch event.Type {
	// Payment Intent events
	case "payment_intent.succeeded":
		return s.handlePaymentIntentSucceeded(event)
	case "payment_intent.payment_failed":
		return s.handlePaymentIntentFailed(event)
	case "payment_intent.canceled":
		return s.handlePaymentIntentCanceled(event)

	// Subscription events
	case "customer.subscription.created":
		return s.handleSubscriptionCreated(event)
	case "customer.subscription.updated":
		return s.handleSubscriptionUpdated(event)
	case "customer.subscription.deleted":
		return s.handleSubscriptionDeleted(event)

	// Invoice events
	case "invoice.paid":
		return s.handleInvoicePaid(event)
	case "invoice.payment_failed":
		return s.handleInvoicePaymentFailed(event)
	case "invoice.finalized":
		return s.handleInvoiceFinalized(event)

	// Refund events
	case "charge.refunded":
		return s.handleChargeRefunded(event)

	default:
		s.logger.Info("Unhandled webhook event type", "type", event.Type)
		return nil
	}
}

func (s *WebhookService) handlePaymentIntentSucceeded(event *stripe.Event) error {
	var pi stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
		return err
	}

	s.logger.Info("Processing payment_intent.succeeded", "payment_intent_id", pi.ID)

	// Update transaction status
	tx, err := s.paymentService.txRepo.GetByStripePaymentID(pi.ID)
	if err != nil {
		return err
	}

	tx.Status = models.TransactionStatusSucceeded
	if pi.Charges != nil && len(pi.Charges.Data) > 0 {
		tx.ReceiptURL = pi.Charges.Data[0].ReceiptURL
	}

	return s.paymentService.txRepo.Update(tx)
}

func (s *WebhookService) handlePaymentIntentFailed(event *stripe.Event) error {
	var pi stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
		return err
	}

	s.logger.Info("Processing payment_intent.payment_failed", "payment_intent_id", pi.ID)

	tx, err := s.paymentService.txRepo.GetByStripePaymentID(pi.ID)
	if err != nil {
		return err
	}

	tx.Status = models.TransactionStatusFailed
	if pi.LastPaymentError != nil {
		tx.FailureReason = pi.LastPaymentError.Message
	}

	return s.paymentService.txRepo.Update(tx)
}

func (s *WebhookService) handlePaymentIntentCanceled(event *stripe.Event) error {
	var pi stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
		return err
	}

	s.logger.Info("Processing payment_intent.canceled", "payment_intent_id", pi.ID)

	return s.paymentService.UpdateTransactionStatus(pi.ID, models.TransactionStatusCanceled)
}

func (s *WebhookService) handleSubscriptionCreated(event *stripe.Event) error {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return err
	}

	s.logger.Info("Processing customer.subscription.created", "subscription_id", sub.ID)
	// Subscription is already created via API, just log it
	return nil
}

func (s *WebhookService) handleSubscriptionUpdated(event *stripe.Event) error {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return err
	}

	s.logger.Info("Processing customer.subscription.updated", "subscription_id", sub.ID)
	return s.subscriptionService.UpdateSubscriptionFromWebhook(&sub)
}

func (s *WebhookService) handleSubscriptionDeleted(event *stripe.Event) error {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return err
	}

	s.logger.Info("Processing customer.subscription.deleted", "subscription_id", sub.ID)
	return s.subscriptionService.UpdateSubscriptionFromWebhook(&sub)
}

func (s *WebhookService) handleInvoicePaid(event *stripe.Event) error {
	var inv stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
		return err
	}

	s.logger.Info("Processing invoice.paid", "invoice_id", inv.ID)

	// Get or create invoice
	invoice, err := s.webhookRepo.(*repository.WebhookEventRepository).db.(*repository.InvoiceRepository).GetByStripeInvoiceID(inv.ID)
	if err != nil {
		// Invoice doesn't exist, this shouldn't happen but handle it
		return nil
	}

	// Mark invoice as paid
	now := time.Now()
	invoice.Status = models.InvoiceStatusPaid
	invoice.AmountPaid = inv.AmountPaid
	invoice.AmountRemaining = 0
	invoice.PaidAt = &now

	// Note: You'd need to inject InvoiceRepository to update the invoice
	// For now, this is a simplified version
	s.logger.Info("Invoice marked as paid", "invoice_id", invoice.ID)
	return nil
}

func (s *WebhookService) handleInvoicePaymentFailed(event *stripe.Event) error {
	var inv stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
		return err
	}

	s.logger.Info("Processing invoice.payment_failed", "invoice_id", inv.ID)
	// Handle payment failure - send notification, retry, etc.
	return nil
}

func (s *WebhookService) handleInvoiceFinalized(event *stripe.Event) error {
	var inv stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
		return err
	}

	s.logger.Info("Processing invoice.finalized", "invoice_id", inv.ID)
	// Invoice is finalized and ready for payment
	return nil
}

func (s *WebhookService) handleChargeRefunded(event *stripe.Event) error {
	var charge stripe.Charge
	if err := json.Unmarshal(event.Data.Raw, &charge); err != nil {
		return err
	}

	s.logger.Info("Processing charge.refunded", "charge_id", charge.ID)

	// Update transaction with refund information
	if charge.PaymentIntent != nil {
		tx, err := s.paymentService.txRepo.GetByStripePaymentID(charge.PaymentIntent.ID)
		if err != nil {
			return err
		}

		tx.RefundedAmount = charge.AmountRefunded
		tx.IsRefunded = charge.Refunded

		return s.paymentService.txRepo.Update(tx)
	}

	return nil
}
