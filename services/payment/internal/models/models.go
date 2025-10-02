package models

import (
	"time"
)

// Transaction represents a payment transaction
type Transaction struct {
	ID                 string    `json:"id" db:"id"`
	UserID             string    `json:"user_id" db:"user_id"`
	StripePaymentID    string    `json:"stripe_payment_id" db:"stripe_payment_id"`
	Amount             int64     `json:"amount" db:"amount"` // Amount in cents
	Currency           string    `json:"currency" db:"currency"`
	Status             string    `json:"status" db:"status"`
	PaymentMethod      string    `json:"payment_method" db:"payment_method"`
	Description        string    `json:"description" db:"description"`
	Metadata           string    `json:"metadata,omitempty" db:"metadata"` // JSON string
	FailureReason      string    `json:"failure_reason,omitempty" db:"failure_reason"`
	RefundedAmount     int64     `json:"refunded_amount" db:"refunded_amount"`
	IsRefunded         bool      `json:"is_refunded" db:"is_refunded"`
	ReceiptURL         string    `json:"receipt_url,omitempty" db:"receipt_url"`
	InvoiceID          *string   `json:"invoice_id,omitempty" db:"invoice_id"`
	SubscriptionID     *string   `json:"subscription_id,omitempty" db:"subscription_id"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// TransactionStatus constants
const (
	TransactionStatusPending   = "pending"
	TransactionStatusSucceeded = "succeeded"
	TransactionStatusFailed    = "failed"
	TransactionStatusRefunded  = "refunded"
	TransactionStatusCanceled  = "canceled"
)

// Subscription represents a recurring subscription
type Subscription struct {
	ID                 string     `json:"id" db:"id"`
	UserID             string     `json:"user_id" db:"user_id"`
	StripeSubscriptionID string   `json:"stripe_subscription_id" db:"stripe_subscription_id"`
	StripePriceID      string     `json:"stripe_price_id" db:"stripe_price_id"`
	StripeProductID    string     `json:"stripe_product_id" db:"stripe_product_id"`
	Status             string     `json:"status" db:"status"`
	PlanName           string     `json:"plan_name" db:"plan_name"`
	Amount             int64      `json:"amount" db:"amount"` // Amount in cents
	Currency           string     `json:"currency" db:"currency"`
	Interval           string     `json:"interval" db:"interval"` // month, year
	IntervalCount      int        `json:"interval_count" db:"interval_count"`
	TrialEnd           *time.Time `json:"trial_end,omitempty" db:"trial_end"`
	CurrentPeriodStart time.Time  `json:"current_period_start" db:"current_period_start"`
	CurrentPeriodEnd   time.Time  `json:"current_period_end" db:"current_period_end"`
	CancelAtPeriodEnd  bool       `json:"cancel_at_period_end" db:"cancel_at_period_end"`
	CanceledAt         *time.Time `json:"canceled_at,omitempty" db:"canceled_at"`
	EndedAt            *time.Time `json:"ended_at,omitempty" db:"ended_at"`
	Metadata           string     `json:"metadata,omitempty" db:"metadata"` // JSON string
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}

// SubscriptionStatus constants
const (
	SubscriptionStatusActive    = "active"
	SubscriptionStatusPastDue   = "past_due"
	SubscriptionStatusCanceled  = "canceled"
	SubscriptionStatusTrialing  = "trialing"
	SubscriptionStatusIncomplete = "incomplete"
)

// Invoice represents a billing invoice
type Invoice struct {
	ID               string     `json:"id" db:"id"`
	UserID           string     `json:"user_id" db:"user_id"`
	StripeInvoiceID  string     `json:"stripe_invoice_id" db:"stripe_invoice_id"`
	SubscriptionID   *string    `json:"subscription_id,omitempty" db:"subscription_id"`
	Number           string     `json:"number" db:"number"`
	Status           string     `json:"status" db:"status"`
	AmountDue        int64      `json:"amount_due" db:"amount_due"`
	AmountPaid       int64      `json:"amount_paid" db:"amount_paid"`
	AmountRemaining  int64      `json:"amount_remaining" db:"amount_remaining"`
	Currency         string     `json:"currency" db:"currency"`
	Description      string     `json:"description,omitempty" db:"description"`
	HostedInvoiceURL string     `json:"hosted_invoice_url,omitempty" db:"hosted_invoice_url"`
	InvoicePDF       string     `json:"invoice_pdf,omitempty" db:"invoice_pdf"`
	DueDate          *time.Time `json:"due_date,omitempty" db:"due_date"`
	PaidAt           *time.Time `json:"paid_at,omitempty" db:"paid_at"`
	VoidedAt         *time.Time `json:"voided_at,omitempty" db:"voided_at"`
	Metadata         string     `json:"metadata,omitempty" db:"metadata"` // JSON string
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// InvoiceStatus constants
const (
	InvoiceStatusDraft         = "draft"
	InvoiceStatusOpen          = "open"
	InvoiceStatusPaid          = "paid"
	InvoiceStatusUncollectible = "uncollectible"
	InvoiceStatusVoid          = "void"
)

// PaymentMethod represents a stored payment method
type PaymentMethod struct {
	ID                   string    `json:"id" db:"id"`
	UserID               string    `json:"user_id" db:"user_id"`
	StripePaymentMethodID string   `json:"stripe_payment_method_id" db:"stripe_payment_method_id"`
	Type                 string    `json:"type" db:"type"` // card, bank_account, etc.
	IsDefault            bool      `json:"is_default" db:"is_default"`
	CardBrand            *string   `json:"card_brand,omitempty" db:"card_brand"`
	CardLast4            *string   `json:"card_last4,omitempty" db:"card_last4"`
	CardExpMonth         *int      `json:"card_exp_month,omitempty" db:"card_exp_month"`
	CardExpYear          *int      `json:"card_exp_year,omitempty" db:"card_exp_year"`
	BankName             *string   `json:"bank_name,omitempty" db:"bank_name"`
	BankLast4            *string   `json:"bank_last4,omitempty" db:"bank_last4"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}

// Refund represents a payment refund
type Refund struct {
	ID              string    `json:"id" db:"id"`
	TransactionID   string    `json:"transaction_id" db:"transaction_id"`
	StripeRefundID  string    `json:"stripe_refund_id" db:"stripe_refund_id"`
	Amount          int64     `json:"amount" db:"amount"` // Amount in cents
	Currency        string    `json:"currency" db:"currency"`
	Reason          string    `json:"reason" db:"reason"`
	Status          string    `json:"status" db:"status"`
	FailureReason   string    `json:"failure_reason,omitempty" db:"failure_reason"`
	Metadata        string    `json:"metadata,omitempty" db:"metadata"` // JSON string
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// RefundReason constants
const (
	RefundReasonDuplicate          = "duplicate"
	RefundReasonFraudulent         = "fraudulent"
	RefundReasonRequestedByCustomer = "requested_by_customer"
)

// RefundStatus constants
const (
	RefundStatusPending   = "pending"
	RefundStatusSucceeded = "succeeded"
	RefundStatusFailed    = "failed"
	RefundStatusCanceled  = "canceled"
)

// WebhookEvent represents a received webhook event
type WebhookEvent struct {
	ID          string    `json:"id" db:"id"`
	StripeEventID string  `json:"stripe_event_id" db:"stripe_event_id"`
	Type        string    `json:"type" db:"type"`
	APIVersion  string    `json:"api_version" db:"api_version"`
	Data        string    `json:"data" db:"data"` // JSON string
	Processed   bool      `json:"processed" db:"processed"`
	ProcessedAt *time.Time `json:"processed_at,omitempty" db:"processed_at"`
	Error       string    `json:"error,omitempty" db:"error"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// PaymentIntent represents a Stripe payment intent
type PaymentIntent struct {
	ID             string                 `json:"id"`
	Amount         int64                  `json:"amount"`
	Currency       string                 `json:"currency"`
	Status         string                 `json:"status"`
	ClientSecret   string                 `json:"client_secret"`
	PaymentMethod  string                 `json:"payment_method,omitempty"`
	Description    string                 `json:"description,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// CreatePaymentRequest represents a payment creation request
type CreatePaymentRequest struct {
	UserID        string                 `json:"user_id" binding:"required"`
	Amount        int64                  `json:"amount" binding:"required,min=50"` // Minimum 50 cents
	Currency      string                 `json:"currency" binding:"required,len=3"`
	PaymentMethod string                 `json:"payment_method,omitempty"`
	Description   string                 `json:"description"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// CreateSubscriptionRequest represents a subscription creation request
type CreateSubscriptionRequest struct {
	UserID        string                 `json:"user_id" binding:"required"`
	PriceID       string                 `json:"price_id" binding:"required"`
	PaymentMethod string                 `json:"payment_method" binding:"required"`
	TrialDays     int                    `json:"trial_days,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// CreateRefundRequest represents a refund creation request
type CreateRefundRequest struct {
	TransactionID string `json:"transaction_id" binding:"required"`
	Amount        int64  `json:"amount,omitempty"` // Partial refund, omit for full refund
	Reason        string `json:"reason" binding:"required"`
}

// PaymentMethodRequest represents a payment method attachment request
type PaymentMethodRequest struct {
	UserID        string `json:"user_id" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required"`
	SetAsDefault  bool   `json:"set_as_default"`
}
