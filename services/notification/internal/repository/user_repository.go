package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/b25/services/notification/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// UserRepository handles database operations for users
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByExternalID(ctx context.Context, externalID string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Preferences
	CreatePreference(ctx context.Context, pref *models.NotificationPreference) error
	GetPreferences(ctx context.Context, userID uuid.UUID) ([]*models.NotificationPreference, error)
	GetPreference(ctx context.Context, userID uuid.UUID, channel models.NotificationChannel, category string) (*models.NotificationPreference, error)
	UpdatePreference(ctx context.Context, pref *models.NotificationPreference) error
	DeletePreference(ctx context.Context, id uuid.UUID) error

	// Devices
	CreateDevice(ctx context.Context, device *models.UserDevice) error
	GetDevices(ctx context.Context, userID uuid.UUID) ([]*models.UserDevice, error)
	GetActiveDevices(ctx context.Context, userID uuid.UUID) ([]*models.UserDevice, error)
	UpdateDevice(ctx context.Context, device *models.UserDevice) error
	DeleteDevice(ctx context.Context, id uuid.UUID) error
	DeactivateDevice(ctx context.Context, id uuid.UUID) error
}

type userRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, external_user_id, email, phone_number, timezone, language)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		user.ID,
		user.ExternalUserID,
		user.Email,
		user.PhoneNumber,
		user.Timezone,
		user.Language,
	).Scan(&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, external_user_id, email, phone_number, timezone, language, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.ExternalUserID,
		&user.Email,
		&user.PhoneNumber,
		&user.Timezone,
		&user.Language,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *userRepository) GetByExternalID(ctx context.Context, externalID string) (*models.User, error) {
	query := `
		SELECT id, external_user_id, email, phone_number, timezone, language, created_at, updated_at, deleted_at
		FROM users
		WHERE external_user_id = $1 AND deleted_at IS NULL
	`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, externalID).Scan(
		&user.ID,
		&user.ExternalUserID,
		&user.Email,
		&user.PhoneNumber,
		&user.Timezone,
		&user.Language,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users SET
			email = $2,
			phone_number = $3,
			timezone = $4,
			language = $5
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(
		ctx, query,
		user.ID,
		user.Email,
		user.PhoneNumber,
		user.Timezone,
		user.Language,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// Preferences methods

func (r *userRepository) CreatePreference(ctx context.Context, pref *models.NotificationPreference) error {
	query := `
		INSERT INTO notification_preferences (
			id, user_id, channel, category, is_enabled,
			quiet_hours_enabled, quiet_hours_start, quiet_hours_end
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id, channel, category)
		DO UPDATE SET
			is_enabled = EXCLUDED.is_enabled,
			quiet_hours_enabled = EXCLUDED.quiet_hours_enabled,
			quiet_hours_start = EXCLUDED.quiet_hours_start,
			quiet_hours_end = EXCLUDED.quiet_hours_end
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		pref.ID,
		pref.UserID,
		pref.Channel,
		pref.Category,
		pref.IsEnabled,
		pref.QuietHoursEnabled,
		pref.QuietHoursStart,
		pref.QuietHoursEnd,
	).Scan(&pref.CreatedAt, &pref.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create preference: %w", err)
	}

	return nil
}

func (r *userRepository) GetPreferences(ctx context.Context, userID uuid.UUID) ([]*models.NotificationPreference, error) {
	query := `
		SELECT
			id, user_id, channel, category, is_enabled,
			quiet_hours_enabled, quiet_hours_start, quiet_hours_end,
			created_at, updated_at
		FROM notification_preferences
		WHERE user_id = $1
		ORDER BY channel, category
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get preferences: %w", err)
	}
	defer rows.Close()

	preferences := make([]*models.NotificationPreference, 0)
	for rows.Next() {
		var pref models.NotificationPreference
		err := rows.Scan(
			&pref.ID,
			&pref.UserID,
			&pref.Channel,
			&pref.Category,
			&pref.IsEnabled,
			&pref.QuietHoursEnabled,
			&pref.QuietHoursStart,
			&pref.QuietHoursEnd,
			&pref.CreatedAt,
			&pref.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan preference: %w", err)
		}
		preferences = append(preferences, &pref)
	}

	return preferences, nil
}

func (r *userRepository) GetPreference(ctx context.Context, userID uuid.UUID, channel models.NotificationChannel, category string) (*models.NotificationPreference, error) {
	query := `
		SELECT
			id, user_id, channel, category, is_enabled,
			quiet_hours_enabled, quiet_hours_start, quiet_hours_end,
			created_at, updated_at
		FROM notification_preferences
		WHERE user_id = $1 AND channel = $2 AND category = $3
	`

	var pref models.NotificationPreference
	err := r.db.QueryRowContext(ctx, query, userID, channel, category).Scan(
		&pref.ID,
		&pref.UserID,
		&pref.Channel,
		&pref.Category,
		&pref.IsEnabled,
		&pref.QuietHoursEnabled,
		&pref.QuietHoursStart,
		&pref.QuietHoursEnd,
		&pref.CreatedAt,
		&pref.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("preference not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get preference: %w", err)
	}

	return &pref, nil
}

func (r *userRepository) UpdatePreference(ctx context.Context, pref *models.NotificationPreference) error {
	query := `
		UPDATE notification_preferences SET
			is_enabled = $2,
			quiet_hours_enabled = $3,
			quiet_hours_start = $4,
			quiet_hours_end = $5
		WHERE id = $1
	`

	result, err := r.db.ExecContext(
		ctx, query,
		pref.ID,
		pref.IsEnabled,
		pref.QuietHoursEnabled,
		pref.QuietHoursStart,
		pref.QuietHoursEnd,
	)

	if err != nil {
		return fmt.Errorf("failed to update preference: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("preference not found")
	}

	return nil
}

func (r *userRepository) DeletePreference(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM notification_preferences WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete preference: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("preference not found")
	}

	return nil
}

// Device methods

func (r *userRepository) CreateDevice(ctx context.Context, device *models.UserDevice) error {
	query := `
		INSERT INTO user_devices (id, user_id, device_token, device_type, device_name, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (device_token)
		DO UPDATE SET
			user_id = EXCLUDED.user_id,
			device_type = EXCLUDED.device_type,
			device_name = EXCLUDED.device_name,
			is_active = EXCLUDED.is_active,
			last_used_at = NOW()
		RETURNING created_at, updated_at, last_used_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		device.ID,
		device.UserID,
		device.DeviceToken,
		device.DeviceType,
		device.DeviceName,
		device.IsActive,
	).Scan(&device.CreatedAt, &device.UpdatedAt, &device.LastUsedAt)

	if err != nil {
		return fmt.Errorf("failed to create device: %w", err)
	}

	return nil
}

func (r *userRepository) GetDevices(ctx context.Context, userID uuid.UUID) ([]*models.UserDevice, error) {
	query := `
		SELECT id, user_id, device_token, device_type, device_name, is_active, last_used_at, created_at, updated_at
		FROM user_devices
		WHERE user_id = $1
		ORDER BY last_used_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}
	defer rows.Close()

	devices := make([]*models.UserDevice, 0)
	for rows.Next() {
		var device models.UserDevice
		err := rows.Scan(
			&device.ID,
			&device.UserID,
			&device.DeviceToken,
			&device.DeviceType,
			&device.DeviceName,
			&device.IsActive,
			&device.LastUsedAt,
			&device.CreatedAt,
			&device.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}
		devices = append(devices, &device)
	}

	return devices, nil
}

func (r *userRepository) GetActiveDevices(ctx context.Context, userID uuid.UUID) ([]*models.UserDevice, error) {
	query := `
		SELECT id, user_id, device_token, device_type, device_name, is_active, last_used_at, created_at, updated_at
		FROM user_devices
		WHERE user_id = $1 AND is_active = true
		ORDER BY last_used_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active devices: %w", err)
	}
	defer rows.Close()

	devices := make([]*models.UserDevice, 0)
	for rows.Next() {
		var device models.UserDevice
		err := rows.Scan(
			&device.ID,
			&device.UserID,
			&device.DeviceToken,
			&device.DeviceType,
			&device.DeviceName,
			&device.IsActive,
			&device.LastUsedAt,
			&device.CreatedAt,
			&device.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}
		devices = append(devices, &device)
	}

	return devices, nil
}

func (r *userRepository) UpdateDevice(ctx context.Context, device *models.UserDevice) error {
	query := `
		UPDATE user_devices SET
			device_name = $2,
			is_active = $3,
			last_used_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.ExecContext(
		ctx, query,
		device.ID,
		device.DeviceName,
		device.IsActive,
	)

	if err != nil {
		return fmt.Errorf("failed to update device: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("device not found")
	}

	return nil
}

func (r *userRepository) DeleteDevice(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM user_devices WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete device: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("device not found")
	}

	return nil
}

func (r *userRepository) DeactivateDevice(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE user_devices SET is_active = false WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to deactivate device: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("device not found")
	}

	return nil
}
