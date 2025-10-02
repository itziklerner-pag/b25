// +build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/b25/services/notification/internal/config"
	"github.com/b25/services/notification/internal/models"
	"github.com/b25/services/notification/internal/repository"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	cfg := config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "notification_user",
		Password: "test_password",
		DBName:   "notification_test_db",
		SSLMode:  "disable",
	}

	db, err := sqlx.Connect("postgres", cfg.GetDSN())
	require.NoError(t, err)

	return db
}

func TestNotificationRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewNotificationRepository(db)
	ctx := context.Background()

	notification := &models.Notification{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		Channel:  models.ChannelEmail,
		Priority: models.PriorityNormal,
		Status:   models.StatusPending,
		Body:     "Test notification body",
		Subject:  stringPtr("Test Subject"),
	}

	err := repo.Create(ctx, notification)
	assert.NoError(t, err)
	assert.NotZero(t, notification.CreatedAt)

	// Verify it was created
	retrieved, err := repo.GetByID(ctx, notification.ID)
	assert.NoError(t, err)
	assert.Equal(t, notification.ID, retrieved.ID)
	assert.Equal(t, notification.Body, retrieved.Body)
}

func TestNotificationRepository_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewNotificationRepository(db)
	ctx := context.Background()

	// Create a notification
	notification := &models.Notification{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		Channel:  models.ChannelEmail,
		Priority: models.PriorityNormal,
		Status:   models.StatusPending,
		Body:     "Test notification",
	}

	err := repo.Create(ctx, notification)
	require.NoError(t, err)

	// Update status
	err = repo.UpdateStatus(ctx, notification.ID, models.StatusSent)
	assert.NoError(t, err)

	// Verify status was updated
	retrieved, err := repo.GetByID(ctx, notification.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.StatusSent, retrieved.Status)
}

func stringPtr(s string) *string {
	return &s
}
