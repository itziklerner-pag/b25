package unit

import (
	"context"
	"testing"

	"github.com/b25/services/messaging/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of the Repository interface
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// Add other mock methods as needed...

func TestCreateUser(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
	}

	mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)

	// Test
	err := mockRepo.CreateUser(context.Background(), user)

	// Assert
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.ID)
	mockRepo.AssertExpectations(t)
}

func TestGetUser(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	userID := uuid.New()
	expectedUser := &models.User{
		ID:       userID,
		Username: "testuser",
		Email:    "test@example.com",
	}

	mockRepo.On("GetUser", mock.Anything, userID).Return(expectedUser, nil)

	// Test
	user, err := mockRepo.GetUser(context.Background(), userID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, "testuser", user.Username)
	mockRepo.AssertExpectations(t)
}
