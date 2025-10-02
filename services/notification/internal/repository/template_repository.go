package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/b25/services/notification/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// TemplateRepository handles database operations for templates
type TemplateRepository interface {
	Create(ctx context.Context, template *models.NotificationTemplate) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.NotificationTemplate, error)
	GetByName(ctx context.Context, name string) (*models.NotificationTemplate, error)
	Update(ctx context.Context, template *models.NotificationTemplate) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, channel *models.NotificationChannel, templateType *models.TemplateType) ([]*models.NotificationTemplate, error)
	ListActive(ctx context.Context) ([]*models.NotificationTemplate, error)
}

type templateRepository struct {
	db *sqlx.DB
}

// NewTemplateRepository creates a new template repository
func NewTemplateRepository(db *sqlx.DB) TemplateRepository {
	return &templateRepository{db: db}
}

func (r *templateRepository) Create(ctx context.Context, template *models.NotificationTemplate) error {
	query := `
		INSERT INTO notification_templates (
			id, name, type, channel, subject, body_template,
			variables, is_active, version, created_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
		RETURNING created_at, updated_at
	`

	variables, err := json.Marshal(template.Variables)
	if err != nil {
		return fmt.Errorf("failed to marshal variables: %w", err)
	}

	err = r.db.QueryRowContext(
		ctx, query,
		template.ID,
		template.Name,
		template.Type,
		template.Channel,
		template.Subject,
		template.BodyTemplate,
		variables,
		template.IsActive,
		template.Version,
		template.CreatedBy,
	).Scan(&template.CreatedAt, &template.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}

	return nil
}

func (r *templateRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.NotificationTemplate, error) {
	query := `
		SELECT
			id, name, type, channel, subject, body_template,
			variables, is_active, version, created_by, created_at, updated_at
		FROM notification_templates
		WHERE id = $1
	`

	var template models.NotificationTemplate
	var variablesJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&template.ID,
		&template.Name,
		&template.Type,
		&template.Channel,
		&template.Subject,
		&template.BodyTemplate,
		&variablesJSON,
		&template.IsActive,
		&template.Version,
		&template.CreatedBy,
		&template.CreatedAt,
		&template.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("template not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	if len(variablesJSON) > 0 {
		if err := json.Unmarshal(variablesJSON, &template.Variables); err != nil {
			return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
		}
	}

	return &template, nil
}

func (r *templateRepository) GetByName(ctx context.Context, name string) (*models.NotificationTemplate, error) {
	query := `
		SELECT
			id, name, type, channel, subject, body_template,
			variables, is_active, version, created_by, created_at, updated_at
		FROM notification_templates
		WHERE name = $1 AND is_active = true
	`

	var template models.NotificationTemplate
	var variablesJSON []byte

	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&template.ID,
		&template.Name,
		&template.Type,
		&template.Channel,
		&template.Subject,
		&template.BodyTemplate,
		&variablesJSON,
		&template.IsActive,
		&template.Version,
		&template.CreatedBy,
		&template.CreatedAt,
		&template.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("template not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	if len(variablesJSON) > 0 {
		if err := json.Unmarshal(variablesJSON, &template.Variables); err != nil {
			return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
		}
	}

	return &template, nil
}

func (r *templateRepository) Update(ctx context.Context, template *models.NotificationTemplate) error {
	query := `
		UPDATE notification_templates SET
			name = $2,
			type = $3,
			channel = $4,
			subject = $5,
			body_template = $6,
			variables = $7,
			is_active = $8,
			version = $9
		WHERE id = $1
	`

	variables, err := json.Marshal(template.Variables)
	if err != nil {
		return fmt.Errorf("failed to marshal variables: %w", err)
	}

	result, err := r.db.ExecContext(
		ctx, query,
		template.ID,
		template.Name,
		template.Type,
		template.Channel,
		template.Subject,
		template.BodyTemplate,
		variables,
		template.IsActive,
		template.Version,
	)

	if err != nil {
		return fmt.Errorf("failed to update template: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("template not found")
	}

	return nil
}

func (r *templateRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM notification_templates WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("template not found")
	}

	return nil
}

func (r *templateRepository) List(ctx context.Context, channel *models.NotificationChannel, templateType *models.TemplateType) ([]*models.NotificationTemplate, error) {
	query := `
		SELECT
			id, name, type, channel, subject, body_template,
			variables, is_active, version, created_by, created_at, updated_at
		FROM notification_templates
		WHERE 1=1
	`

	args := []interface{}{}
	argIndex := 1

	if channel != nil {
		query += fmt.Sprintf(" AND channel = $%d", argIndex)
		args = append(args, *channel)
		argIndex++
	}

	if templateType != nil {
		query += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, *templateType)
		argIndex++
	}

	query += " ORDER BY name ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}
	defer rows.Close()

	templates := make([]*models.NotificationTemplate, 0)
	for rows.Next() {
		var template models.NotificationTemplate
		var variablesJSON []byte

		err := rows.Scan(
			&template.ID,
			&template.Name,
			&template.Type,
			&template.Channel,
			&template.Subject,
			&template.BodyTemplate,
			&variablesJSON,
			&template.IsActive,
			&template.Version,
			&template.CreatedBy,
			&template.CreatedAt,
			&template.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template: %w", err)
		}

		if len(variablesJSON) > 0 {
			if err := json.Unmarshal(variablesJSON, &template.Variables); err != nil {
				return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
			}
		}

		templates = append(templates, &template)
	}

	return templates, nil
}

func (r *templateRepository) ListActive(ctx context.Context) ([]*models.NotificationTemplate, error) {
	query := `
		SELECT
			id, name, type, channel, subject, body_template,
			variables, is_active, version, created_by, created_at, updated_at
		FROM notification_templates
		WHERE is_active = true
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list active templates: %w", err)
	}
	defer rows.Close()

	templates := make([]*models.NotificationTemplate, 0)
	for rows.Next() {
		var template models.NotificationTemplate
		var variablesJSON []byte

		err := rows.Scan(
			&template.ID,
			&template.Name,
			&template.Type,
			&template.Channel,
			&template.Subject,
			&template.BodyTemplate,
			&variablesJSON,
			&template.IsActive,
			&template.Version,
			&template.CreatedBy,
			&template.CreatedAt,
			&template.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template: %w", err)
		}

		if len(variablesJSON) > 0 {
			if err := json.Unmarshal(variablesJSON, &template.Variables); err != nil {
				return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
			}
		}

		templates = append(templates, &template)
	}

	return templates, nil
}
