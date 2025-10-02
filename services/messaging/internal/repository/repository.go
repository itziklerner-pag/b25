package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/b25/services/messaging/internal/models"
	"github.com/google/uuid"
)

// Repository defines the interface for data access
type Repository interface {
	// User operations
	CreateUser(ctx context.Context, user *models.User) error
	GetUser(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) error
	UpdateUserLastSeen(ctx context.Context, userID uuid.UUID, lastSeen time.Time) error

	// Conversation operations
	CreateConversation(ctx context.Context, conversation *models.Conversation) error
	GetConversation(ctx context.Context, id uuid.UUID) (*models.Conversation, error)
	GetConversationWithDetails(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.ConversationWithDetails, error)
	ListUserConversations(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.ConversationWithDetails, error)
	UpdateConversation(ctx context.Context, conversation *models.Conversation) error
	DeleteConversation(ctx context.Context, id uuid.UUID) error
	GetDirectConversation(ctx context.Context, user1ID, user2ID uuid.UUID) (*models.Conversation, error)

	// Conversation member operations
	AddConversationMember(ctx context.Context, member *models.ConversationMember) error
	RemoveConversationMember(ctx context.Context, conversationID, userID uuid.UUID) error
	GetConversationMembers(ctx context.Context, conversationID uuid.UUID) ([]models.User, error)
	IsConversationMember(ctx context.Context, conversationID, userID uuid.UUID) (bool, error)
	UpdateMemberLastRead(ctx context.Context, conversationID, userID uuid.UUID, lastReadAt time.Time) error
	MuteConversation(ctx context.Context, conversationID, userID uuid.UUID, muted bool) error

	// Message operations
	CreateMessage(ctx context.Context, message *models.Message) error
	GetMessage(ctx context.Context, id uuid.UUID) (*models.Message, error)
	GetMessageWithDetails(ctx context.Context, id uuid.UUID) (*models.MessageWithDetails, error)
	ListConversationMessages(ctx context.Context, conversationID uuid.UUID, limit, offset int) ([]models.MessageWithDetails, error)
	UpdateMessage(ctx context.Context, message *models.Message) error
	DeleteMessage(ctx context.Context, id uuid.UUID) error
	SoftDeleteMessage(ctx context.Context, id uuid.UUID) error
	SearchMessages(ctx context.Context, userID uuid.UUID, query string, limit, offset int) ([]models.MessageWithDetails, error)

	// Message reaction operations
	AddReaction(ctx context.Context, reaction *models.MessageReaction) error
	RemoveReaction(ctx context.Context, messageID, userID uuid.UUID, emoji string) error
	GetMessageReactions(ctx context.Context, messageID uuid.UUID) ([]models.MessageReaction, error)

	// Message read receipt operations
	MarkMessageAsRead(ctx context.Context, receipt *models.MessageReadReceipt) error
	GetMessageReadReceipts(ctx context.Context, messageID uuid.UUID) ([]models.MessageReadReceipt, error)
	GetUnreadCount(ctx context.Context, conversationID, userID uuid.UUID) (int, error)

	// File operations
	CreateFile(ctx context.Context, file *models.File) error
	GetFile(ctx context.Context, id uuid.UUID) (*models.File, error)
	GetMessageFiles(ctx context.Context, messageID uuid.UUID) ([]models.File, error)
	DeleteFile(ctx context.Context, id uuid.UUID) error

	// Typing indicator operations
	SetTypingIndicator(ctx context.Context, indicator *models.TypingIndicator) error
	GetTypingIndicators(ctx context.Context, conversationID uuid.UUID) ([]models.TypingIndicator, error)
	RemoveTypingIndicator(ctx context.Context, conversationID, userID uuid.UUID) error
	CleanupExpiredTypingIndicators(ctx context.Context) error

	// Utility
	BeginTx(ctx context.Context) (*sql.Tx, error)
	CommitTx(tx *sql.Tx) error
	RollbackTx(tx *sql.Tx) error
}
