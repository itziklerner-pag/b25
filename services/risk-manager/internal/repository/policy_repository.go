package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/b25/services/risk-manager/internal/limits"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// PolicyRepository handles policy database operations
type PolicyRepository struct {
	db *sqlx.DB
}

// NewPolicyRepository creates a new policy repository
func NewPolicyRepository(db *sqlx.DB) *PolicyRepository {
	return &PolicyRepository{db: db}
}

// policyRow represents the database row structure
type policyRow struct {
	ID        string         `db:"id"`
	Name      string         `db:"name"`
	Type      string         `db:"type"`
	Metric    string         `db:"metric"`
	Operator  string         `db:"operator"`
	Threshold float64        `db:"threshold"`
	Scope     string         `db:"scope"`
	ScopeID   sql.NullString `db:"scope_id"`
	Action    sql.NullString `db:"action"`
	Enabled   bool           `db:"enabled"`
	Priority  int            `db:"priority"`
	Metadata  []byte         `db:"metadata"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
	Version   int            `db:"version"`
}

// GetAll retrieves all policies
func (r *PolicyRepository) GetAll(ctx context.Context) ([]*limits.Policy, error) {
	query := `
		SELECT id, name, type, metric, operator, threshold, scope, scope_id,
		       action, enabled, priority, metadata, created_at, updated_at, version
		FROM risk_policies
		ORDER BY priority DESC, created_at ASC
	`

	var rows []policyRow
	if err := r.db.SelectContext(ctx, &rows, query); err != nil {
		return nil, err
	}

	return r.rowsToPolicies(rows)
}

// GetActive retrieves all active policies
func (r *PolicyRepository) GetActive(ctx context.Context) ([]*limits.Policy, error) {
	query := `
		SELECT id, name, type, metric, operator, threshold, scope, scope_id,
		       action, enabled, priority, metadata, created_at, updated_at, version
		FROM risk_policies
		WHERE enabled = true
		ORDER BY priority DESC, created_at ASC
	`

	var rows []policyRow
	if err := r.db.SelectContext(ctx, &rows, query); err != nil {
		return nil, err
	}

	return r.rowsToPolicies(rows)
}

// GetByID retrieves a policy by ID
func (r *PolicyRepository) GetByID(ctx context.Context, id string) (*limits.Policy, error) {
	query := `
		SELECT id, name, type, metric, operator, threshold, scope, scope_id,
		       action, enabled, priority, metadata, created_at, updated_at, version
		FROM risk_policies
		WHERE id = $1
	`

	var row policyRow
	if err := r.db.GetContext(ctx, &row, query, id); err != nil {
		return nil, err
	}

	policies, err := r.rowsToPolicies([]policyRow{row})
	if err != nil {
		return nil, err
	}

	return policies[0], nil
}

// Create creates a new policy
func (r *PolicyRepository) Create(ctx context.Context, policy *limits.Policy) error {
	if policy.ID == "" {
		policy.ID = uuid.New().String()
	}

	metadata, err := json.Marshal(policy.Metadata)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO risk_policies (
			id, name, type, metric, operator, threshold, scope, scope_id,
			action, enabled, priority, metadata, version
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
		RETURNING created_at, updated_at
	`

	var createdAt, updatedAt time.Time
	err = r.db.QueryRowContext(ctx, query,
		policy.ID, policy.Name, policy.Type, policy.Metric, policy.Operator,
		policy.Threshold, policy.Scope, nullString(policy.ScopeID),
		nullString(policy.Action), policy.Enabled, policy.Priority,
		metadata, 1,
	).Scan(&createdAt, &updatedAt)

	if err != nil {
		return err
	}

	policy.CreatedAt = createdAt
	policy.UpdatedAt = updatedAt
	policy.Version = 1

	return nil
}

// Update updates an existing policy
func (r *PolicyRepository) Update(ctx context.Context, policy *limits.Policy) error {
	metadata, err := json.Marshal(policy.Metadata)
	if err != nil {
		return err
	}

	query := `
		UPDATE risk_policies
		SET name = $1, type = $2, metric = $3, operator = $4, threshold = $5,
		    scope = $6, scope_id = $7, action = $8, enabled = $9, priority = $10,
		    metadata = $11, updated_at = NOW(), version = version + 1
		WHERE id = $12 AND version = $13
		RETURNING updated_at, version
	`

	var updatedAt time.Time
	var version int

	err = r.db.QueryRowContext(ctx, query,
		policy.Name, policy.Type, policy.Metric, policy.Operator, policy.Threshold,
		policy.Scope, nullString(policy.ScopeID), nullString(policy.Action),
		policy.Enabled, policy.Priority, metadata, policy.ID, policy.Version,
	).Scan(&updatedAt, &version)

	if err == sql.ErrNoRows {
		return sql.ErrNoRows // Optimistic locking failure
	}
	if err != nil {
		return err
	}

	policy.UpdatedAt = updatedAt
	policy.Version = version

	return nil
}

// Delete deletes a policy
func (r *PolicyRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM risk_policies WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// RecordViolation records a policy violation
func (r *PolicyRepository) RecordViolation(ctx context.Context, policyID string, metricValue, thresholdValue float64, contextData map[string]interface{}, action string) error {
	contextJSON, err := json.Marshal(contextData)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO risk_violations (
			policy_id, metric_value, threshold_value, context, action_taken
		) VALUES ($1, $2, $3, $4, $5)
	`

	_, err = r.db.ExecContext(ctx, query, policyID, metricValue, thresholdValue, contextJSON, action)
	return err
}

// RecordEmergencyStop records an emergency stop event
func (r *PolicyRepository) RecordEmergencyStop(ctx context.Context, reason, triggeredBy string, accountState, positionsSnapshot map[string]interface{}) (int64, error) {
	accountJSON, err := json.Marshal(accountState)
	if err != nil {
		return 0, err
	}

	positionsJSON, err := json.Marshal(positionsSnapshot)
	if err != nil {
		return 0, err
	}

	query := `
		INSERT INTO emergency_stops (
			trigger_reason, triggered_by, account_state, positions_snapshot
		) VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var id int64
	err = r.db.QueryRowContext(ctx, query, reason, triggeredBy, accountJSON, positionsJSON).Scan(&id)
	return id, err
}

// UpdateEmergencyStop updates emergency stop progress
func (r *PolicyRepository) UpdateEmergencyStop(ctx context.Context, id int64, ordersCanceled, positionsClosed int, completed bool) error {
	query := `
		UPDATE emergency_stops
		SET orders_canceled = $1, positions_closed = $2, completed_at = CASE WHEN $3 THEN NOW() ELSE NULL END
		WHERE id = $4
	`

	_, err := r.db.ExecContext(ctx, query, ordersCanceled, positionsClosed, completed, id)
	return err
}

// rowsToPolicies converts database rows to Policy objects
func (r *PolicyRepository) rowsToPolicies(rows []policyRow) ([]*limits.Policy, error) {
	policies := make([]*limits.Policy, len(rows))

	for i, row := range rows {
		var metadata map[string]interface{}
		if len(row.Metadata) > 0 {
			if err := json.Unmarshal(row.Metadata, &metadata); err != nil {
				return nil, err
			}
		}

		policies[i] = &limits.Policy{
			ID:        row.ID,
			Name:      row.Name,
			Type:      limits.PolicyType(row.Type),
			Metric:    row.Metric,
			Operator:  limits.PolicyOperator(row.Operator),
			Threshold: row.Threshold,
			Scope:     limits.PolicyScope(row.Scope),
			ScopeID:   row.ScopeID.String,
			Action:    row.Action.String,
			Enabled:   row.Enabled,
			Priority:  row.Priority,
			Metadata:  metadata,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
			Version:   row.Version,
		}
	}

	return policies, nil
}

// nullString converts a string to sql.NullString
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
