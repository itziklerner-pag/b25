package payment

import (
	"fmt"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/invoice"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/paymentmethod"
	"github.com/stripe/stripe-go/v76/refund"
	"github.com/stripe/stripe-go/v76/sub"
)

type StripeClient struct {
	apiKey string
}

func NewStripeClient(apiKey string) *StripeClient {
	stripe.Key = apiKey
	return &StripeClient{apiKey: apiKey}
}

// CreatePaymentIntent creates a new payment intent
func (s *StripeClient) CreatePaymentIntent(params *stripe.PaymentIntentParams) (*stripe.PaymentIntent, error) {
	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}
	return pi, nil
}

// GetPaymentIntent retrieves a payment intent
func (s *StripeClient) GetPaymentIntent(id string) (*stripe.PaymentIntent, error) {
	pi, err := paymentintent.Get(id, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment intent: %w", err)
	}
	return pi, nil
}

// ConfirmPaymentIntent confirms a payment intent
func (s *StripeClient) ConfirmPaymentIntent(id string, params *stripe.PaymentIntentConfirmParams) (*stripe.PaymentIntent, error) {
	pi, err := paymentintent.Confirm(id, params)
	if err != nil {
		return nil, fmt.Errorf("failed to confirm payment intent: %w", err)
	}
	return pi, nil
}

// CancelPaymentIntent cancels a payment intent
func (s *StripeClient) CancelPaymentIntent(id string) (*stripe.PaymentIntent, error) {
	pi, err := paymentintent.Cancel(id, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel payment intent: %w", err)
	}
	return pi, nil
}

// CreateCustomer creates a new Stripe customer
func (s *StripeClient) CreateCustomer(params *stripe.CustomerParams) (*stripe.Customer, error) {
	cust, err := customer.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}
	return cust, nil
}

// GetCustomer retrieves a customer
func (s *StripeClient) GetCustomer(id string) (*stripe.Customer, error) {
	cust, err := customer.Get(id, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}
	return cust, nil
}

// AttachPaymentMethod attaches a payment method to a customer
func (s *StripeClient) AttachPaymentMethod(pmID, customerID string) (*stripe.PaymentMethod, error) {
	params := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(customerID),
	}
	pm, err := paymentmethod.Attach(pmID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to attach payment method: %w", err)
	}
	return pm, nil
}

// DetachPaymentMethod detaches a payment method from a customer
func (s *StripeClient) DetachPaymentMethod(pmID string) (*stripe.PaymentMethod, error) {
	pm, err := paymentmethod.Detach(pmID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to detach payment method: %w", err)
	}
	return pm, nil
}

// GetPaymentMethod retrieves a payment method
func (s *StripeClient) GetPaymentMethod(id string) (*stripe.PaymentMethod, error) {
	pm, err := paymentmethod.Get(id, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment method: %w", err)
	}
	return pm, nil
}

// CreateSubscription creates a new subscription
func (s *StripeClient) CreateSubscription(params *stripe.SubscriptionParams) (*stripe.Subscription, error) {
	subscription, err := sub.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}
	return subscription, nil
}

// GetSubscription retrieves a subscription
func (s *StripeClient) GetSubscription(id string) (*stripe.Subscription, error) {
	subscription, err := sub.Get(id, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	return subscription, nil
}

// UpdateSubscription updates a subscription
func (s *StripeClient) UpdateSubscription(id string, params *stripe.SubscriptionParams) (*stripe.Subscription, error) {
	subscription, err := sub.Update(id, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}
	return subscription, nil
}

// CancelSubscription cancels a subscription
func (s *StripeClient) CancelSubscription(id string) (*stripe.Subscription, error) {
	subscription, err := sub.Cancel(id, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel subscription: %w", err)
	}
	return subscription, nil
}

// GetInvoice retrieves an invoice
func (s *StripeClient) GetInvoice(id string) (*stripe.Invoice, error) {
	inv, err := invoice.Get(id, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}
	return inv, nil
}

// FinalizeInvoice finalizes an invoice
func (s *StripeClient) FinalizeInvoice(id string) (*stripe.Invoice, error) {
	inv, err := invoice.FinalizeInvoice(id, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to finalize invoice: %w", err)
	}
	return inv, nil
}

// PayInvoice pays an invoice
func (s *StripeClient) PayInvoice(id string) (*stripe.Invoice, error) {
	inv, err := invoice.Pay(id, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to pay invoice: %w", err)
	}
	return inv, nil
}

// VoidInvoice voids an invoice
func (s *StripeClient) VoidInvoice(id string) (*stripe.Invoice, error) {
	inv, err := invoice.VoidInvoice(id, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to void invoice: %w", err)
	}
	return inv, nil
}

// CreateRefund creates a refund
func (s *StripeClient) CreateRefund(params *stripe.RefundParams) (*stripe.Refund, error) {
	ref, err := refund.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}
	return ref, nil
}

// GetRefund retrieves a refund
func (s *StripeClient) GetRefund(id string) (*stripe.Refund, error) {
	ref, err := refund.Get(id, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get refund: %w", err)
	}
	return ref, nil
}
