package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/yourorg/b25/services/payment/internal/database"
	"github.com/yourorg/b25/services/payment/internal/models"
)

type TransactionRepository struct {
	db *database.DB
}

func NewTransactionRepository(db *database.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(tx *models.Transaction) error {
	metadataJSON, _ := json.Marshal(tx.Metadata)

	query := `
		INSERT INTO transactions (
			id, user_id, stripe_payment_id, amount, currency, status,
			payment_method, description, metadata, receipt_url,
			invoice_id, subscription_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.Exec(
		query,
		tx.ID, tx.UserID, tx.StripePaymentID, tx.Amount, tx.Currency, tx.Status,
		tx.PaymentMethod, tx.Description, metadataJSON, tx.ReceiptURL,
		tx.InvoiceID, tx.SubscriptionID,
	)

	return err
}

func (r *TransactionRepository) GetByID(id string) (*models.Transaction, error) {
	query := `
		SELECT id, user_id, stripe_payment_id, amount, currency, status,
		       payment_method, description, metadata, failure_reason,
		       refunded_amount, is_refunded, receipt_url, invoice_id,
		       subscription_id, created_at, updated_at
		FROM transactions
		WHERE id = $1
	`

	tx := &models.Transaction{}
	var metadata sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&tx.ID, &tx.UserID, &tx.StripePaymentID, &tx.Amount, &tx.Currency,
		&tx.Status, &tx.PaymentMethod, &tx.Description, &metadata,
		&tx.FailureReason, &tx.RefundedAmount, &tx.IsRefunded,
		&tx.ReceiptURL, &tx.InvoiceID, &tx.SubscriptionID,
		&tx.CreatedAt, &tx.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaction not found")
		}
		return nil, err
	}

	if metadata.Valid {
		tx.Metadata = metadata.String
	}

	return tx, nil
}

func (r *TransactionRepository) GetByStripePaymentID(stripePaymentID string) (*models.Transaction, error) {
	query := `
		SELECT id, user_id, stripe_payment_id, amount, currency, status,
		       payment_method, description, metadata, failure_reason,
		       refunded_amount, is_refunded, receipt_url, invoice_id,
		       subscription_id, created_at, updated_at
		FROM transactions
		WHERE stripe_payment_id = $1
	`

	tx := &models.Transaction{}
	var metadata sql.NullString

	err := r.db.QueryRow(query, stripePaymentID).Scan(
		&tx.ID, &tx.UserID, &tx.StripePaymentID, &tx.Amount, &tx.Currency,
		&tx.Status, &tx.PaymentMethod, &tx.Description, &metadata,
		&tx.FailureReason, &tx.RefundedAmount, &tx.IsRefunded,
		&tx.ReceiptURL, &tx.InvoiceID, &tx.SubscriptionID,
		&tx.CreatedAt, &tx.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaction not found")
		}
		return nil, err
	}

	if metadata.Valid {
		tx.Metadata = metadata.String
	}

	return tx, nil
}

func (r *TransactionRepository) GetByUserID(userID string, limit, offset int) ([]*models.Transaction, error) {
	query := `
		SELECT id, user_id, stripe_payment_id, amount, currency, status,
		       payment_method, description, metadata, failure_reason,
		       refunded_amount, is_refunded, receipt_url, invoice_id,
		       subscription_id, created_at, updated_at
		FROM transactions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*models.Transaction

	for rows.Next() {
		tx := &models.Transaction{}
		var metadata sql.NullString

		err := rows.Scan(
			&tx.ID, &tx.UserID, &tx.StripePaymentID, &tx.Amount, &tx.Currency,
			&tx.Status, &tx.PaymentMethod, &tx.Description, &metadata,
			&tx.FailureReason, &tx.RefundedAmount, &tx.IsRefunded,
			&tx.ReceiptURL, &tx.InvoiceID, &tx.SubscriptionID,
			&tx.CreatedAt, &tx.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if metadata.Valid {
			tx.Metadata = metadata.String
		}

		transactions = append(transactions, tx)
	}

	return transactions, nil
}

func (r *TransactionRepository) Update(tx *models.Transaction) error {
	metadataJSON, _ := json.Marshal(tx.Metadata)

	query := `
		UPDATE transactions
		SET status = $1, failure_reason = $2, refunded_amount = $3,
		    is_refunded = $4, receipt_url = $5, metadata = $6
		WHERE id = $7
	`

	_, err := r.db.Exec(
		query,
		tx.Status, tx.FailureReason, tx.RefundedAmount,
		tx.IsRefunded, tx.ReceiptURL, metadataJSON, tx.ID,
	)

	return err
}

func (r *TransactionRepository) UpdateStatus(id, status string) error {
	query := `UPDATE transactions SET status = $1 WHERE id = $2`
	_, err := r.db.Exec(query, status, id)
	return err
}

func (r *TransactionRepository) MarkAsRefunded(id string, refundedAmount int64) error {
	query := `
		UPDATE transactions
		SET refunded_amount = $1, is_refunded = TRUE, status = 'refunded'
		WHERE id = $2
	`
	_, err := r.db.Exec(query, refundedAmount, id)
	return err
}
