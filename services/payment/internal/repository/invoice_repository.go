package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/yourorg/b25/services/payment/internal/database"
	"github.com/yourorg/b25/services/payment/internal/models"
)

type InvoiceRepository struct {
	db *database.DB
}

func NewInvoiceRepository(db *database.DB) *InvoiceRepository {
	return &InvoiceRepository{db: db}
}

func (r *InvoiceRepository) Create(invoice *models.Invoice) error {
	metadataJSON, _ := json.Marshal(invoice.Metadata)

	query := `
		INSERT INTO invoices (
			id, user_id, stripe_invoice_id, subscription_id, number,
			status, amount_due, amount_paid, amount_remaining, currency,
			description, hosted_invoice_url, invoice_pdf, due_date, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err := r.db.Exec(
		query,
		invoice.ID, invoice.UserID, invoice.StripeInvoiceID, invoice.SubscriptionID,
		invoice.Number, invoice.Status, invoice.AmountDue, invoice.AmountPaid,
		invoice.AmountRemaining, invoice.Currency, invoice.Description,
		invoice.HostedInvoiceURL, invoice.InvoicePDF, invoice.DueDate, metadataJSON,
	)

	return err
}

func (r *InvoiceRepository) GetByID(id string) (*models.Invoice, error) {
	query := `
		SELECT id, user_id, stripe_invoice_id, subscription_id, number,
		       status, amount_due, amount_paid, amount_remaining, currency,
		       description, hosted_invoice_url, invoice_pdf, due_date,
		       paid_at, voided_at, metadata, created_at, updated_at
		FROM invoices
		WHERE id = $1
	`

	invoice := &models.Invoice{}
	var metadata sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&invoice.ID, &invoice.UserID, &invoice.StripeInvoiceID,
		&invoice.SubscriptionID, &invoice.Number, &invoice.Status,
		&invoice.AmountDue, &invoice.AmountPaid, &invoice.AmountRemaining,
		&invoice.Currency, &invoice.Description, &invoice.HostedInvoiceURL,
		&invoice.InvoicePDF, &invoice.DueDate, &invoice.PaidAt,
		&invoice.VoidedAt, &metadata, &invoice.CreatedAt, &invoice.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invoice not found")
		}
		return nil, err
	}

	if metadata.Valid {
		invoice.Metadata = metadata.String
	}

	return invoice, nil
}

func (r *InvoiceRepository) GetByStripeInvoiceID(stripeInvoiceID string) (*models.Invoice, error) {
	query := `
		SELECT id, user_id, stripe_invoice_id, subscription_id, number,
		       status, amount_due, amount_paid, amount_remaining, currency,
		       description, hosted_invoice_url, invoice_pdf, due_date,
		       paid_at, voided_at, metadata, created_at, updated_at
		FROM invoices
		WHERE stripe_invoice_id = $1
	`

	invoice := &models.Invoice{}
	var metadata sql.NullString

	err := r.db.QueryRow(query, stripeInvoiceID).Scan(
		&invoice.ID, &invoice.UserID, &invoice.StripeInvoiceID,
		&invoice.SubscriptionID, &invoice.Number, &invoice.Status,
		&invoice.AmountDue, &invoice.AmountPaid, &invoice.AmountRemaining,
		&invoice.Currency, &invoice.Description, &invoice.HostedInvoiceURL,
		&invoice.InvoicePDF, &invoice.DueDate, &invoice.PaidAt,
		&invoice.VoidedAt, &metadata, &invoice.CreatedAt, &invoice.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invoice not found")
		}
		return nil, err
	}

	if metadata.Valid {
		invoice.Metadata = metadata.String
	}

	return invoice, nil
}

func (r *InvoiceRepository) GetByUserID(userID string, limit, offset int) ([]*models.Invoice, error) {
	query := `
		SELECT id, user_id, stripe_invoice_id, subscription_id, number,
		       status, amount_due, amount_paid, amount_remaining, currency,
		       description, hosted_invoice_url, invoice_pdf, due_date,
		       paid_at, voided_at, metadata, created_at, updated_at
		FROM invoices
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []*models.Invoice

	for rows.Next() {
		invoice := &models.Invoice{}
		var metadata sql.NullString

		err := rows.Scan(
			&invoice.ID, &invoice.UserID, &invoice.StripeInvoiceID,
			&invoice.SubscriptionID, &invoice.Number, &invoice.Status,
			&invoice.AmountDue, &invoice.AmountPaid, &invoice.AmountRemaining,
			&invoice.Currency, &invoice.Description, &invoice.HostedInvoiceURL,
			&invoice.InvoicePDF, &invoice.DueDate, &invoice.PaidAt,
			&invoice.VoidedAt, &metadata, &invoice.CreatedAt, &invoice.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if metadata.Valid {
			invoice.Metadata = metadata.String
		}

		invoices = append(invoices, invoice)
	}

	return invoices, nil
}

func (r *InvoiceRepository) GetBySubscriptionID(subscriptionID string) ([]*models.Invoice, error) {
	query := `
		SELECT id, user_id, stripe_invoice_id, subscription_id, number,
		       status, amount_due, amount_paid, amount_remaining, currency,
		       description, hosted_invoice_url, invoice_pdf, due_date,
		       paid_at, voided_at, metadata, created_at, updated_at
		FROM invoices
		WHERE subscription_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, subscriptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []*models.Invoice

	for rows.Next() {
		invoice := &models.Invoice{}
		var metadata sql.NullString

		err := rows.Scan(
			&invoice.ID, &invoice.UserID, &invoice.StripeInvoiceID,
			&invoice.SubscriptionID, &invoice.Number, &invoice.Status,
			&invoice.AmountDue, &invoice.AmountPaid, &invoice.AmountRemaining,
			&invoice.Currency, &invoice.Description, &invoice.HostedInvoiceURL,
			&invoice.InvoicePDF, &invoice.DueDate, &invoice.PaidAt,
			&invoice.VoidedAt, &metadata, &invoice.CreatedAt, &invoice.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if metadata.Valid {
			invoice.Metadata = metadata.String
		}

		invoices = append(invoices, invoice)
	}

	return invoices, nil
}

func (r *InvoiceRepository) Update(invoice *models.Invoice) error {
	metadataJSON, _ := json.Marshal(invoice.Metadata)

	query := `
		UPDATE invoices
		SET status = $1, amount_paid = $2, amount_remaining = $3,
		    paid_at = $4, voided_at = $5, metadata = $6
		WHERE id = $7
	`

	_, err := r.db.Exec(
		query,
		invoice.Status, invoice.AmountPaid, invoice.AmountRemaining,
		invoice.PaidAt, invoice.VoidedAt, metadataJSON, invoice.ID,
	)

	return err
}

func (r *InvoiceRepository) MarkAsPaid(id string) error {
	query := `
		UPDATE invoices
		SET status = 'paid', amount_paid = amount_due,
		    amount_remaining = 0, paid_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(query, id)
	return err
}
