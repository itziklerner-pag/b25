package repository

import (
	"database/sql"
	"fmt"

	"github.com/yourorg/b25/services/payment/internal/database"
	"github.com/yourorg/b25/services/payment/internal/models"
)

type PaymentMethodRepository struct {
	db *database.DB
}

func NewPaymentMethodRepository(db *database.DB) *PaymentMethodRepository {
	return &PaymentMethodRepository{db: db}
}

func (r *PaymentMethodRepository) Create(pm *models.PaymentMethod) error {
	query := `
		INSERT INTO payment_methods (
			id, user_id, stripe_payment_method_id, type, is_default,
			card_brand, card_last4, card_exp_month, card_exp_year,
			bank_name, bank_last4
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.Exec(
		query,
		pm.ID, pm.UserID, pm.StripePaymentMethodID, pm.Type, pm.IsDefault,
		pm.CardBrand, pm.CardLast4, pm.CardExpMonth, pm.CardExpYear,
		pm.BankName, pm.BankLast4,
	)

	return err
}

func (r *PaymentMethodRepository) GetByID(id string) (*models.PaymentMethod, error) {
	query := `
		SELECT id, user_id, stripe_payment_method_id, type, is_default,
		       card_brand, card_last4, card_exp_month, card_exp_year,
		       bank_name, bank_last4, created_at, updated_at
		FROM payment_methods
		WHERE id = $1
	`

	pm := &models.PaymentMethod{}

	err := r.db.QueryRow(query, id).Scan(
		&pm.ID, &pm.UserID, &pm.StripePaymentMethodID, &pm.Type,
		&pm.IsDefault, &pm.CardBrand, &pm.CardLast4, &pm.CardExpMonth,
		&pm.CardExpYear, &pm.BankName, &pm.BankLast4,
		&pm.CreatedAt, &pm.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("payment method not found")
		}
		return nil, err
	}

	return pm, nil
}

func (r *PaymentMethodRepository) GetByUserID(userID string) ([]*models.PaymentMethod, error) {
	query := `
		SELECT id, user_id, stripe_payment_method_id, type, is_default,
		       card_brand, card_last4, card_exp_month, card_exp_year,
		       bank_name, bank_last4, created_at, updated_at
		FROM payment_methods
		WHERE user_id = $1
		ORDER BY is_default DESC, created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paymentMethods []*models.PaymentMethod

	for rows.Next() {
		pm := &models.PaymentMethod{}

		err := rows.Scan(
			&pm.ID, &pm.UserID, &pm.StripePaymentMethodID, &pm.Type,
			&pm.IsDefault, &pm.CardBrand, &pm.CardLast4, &pm.CardExpMonth,
			&pm.CardExpYear, &pm.BankName, &pm.BankLast4,
			&pm.CreatedAt, &pm.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		paymentMethods = append(paymentMethods, pm)
	}

	return paymentMethods, nil
}

func (r *PaymentMethodRepository) GetDefaultByUserID(userID string) (*models.PaymentMethod, error) {
	query := `
		SELECT id, user_id, stripe_payment_method_id, type, is_default,
		       card_brand, card_last4, card_exp_month, card_exp_year,
		       bank_name, bank_last4, created_at, updated_at
		FROM payment_methods
		WHERE user_id = $1 AND is_default = TRUE
		LIMIT 1
	`

	pm := &models.PaymentMethod{}

	err := r.db.QueryRow(query, userID).Scan(
		&pm.ID, &pm.UserID, &pm.StripePaymentMethodID, &pm.Type,
		&pm.IsDefault, &pm.CardBrand, &pm.CardLast4, &pm.CardExpMonth,
		&pm.CardExpYear, &pm.BankName, &pm.BankLast4,
		&pm.CreatedAt, &pm.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no default payment method found")
		}
		return nil, err
	}

	return pm, nil
}

func (r *PaymentMethodRepository) SetAsDefault(id, userID string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Unset all other default payment methods for this user
	_, err = tx.Exec(
		`UPDATE payment_methods SET is_default = FALSE WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return err
	}

	// Set the new default
	_, err = tx.Exec(
		`UPDATE payment_methods SET is_default = TRUE WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PaymentMethodRepository) Delete(id string) error {
	query := `DELETE FROM payment_methods WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
