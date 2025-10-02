package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Username    string     `json:"username" db:"username"`
	Email       string     `json:"email" db:"email"`
	DisplayName *string    `json:"display_name,omitempty" db:"display_name"`
	AvatarURL   *string    `json:"avatar_url,omitempty" db:"avatar_url"`
	Status      string     `json:"status" db:"status"`
	LastSeen    *time.Time `json:"last_seen,omitempty" db:"last_seen"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// ConversationType represents the type of conversation
type ConversationType string

const (
	ConversationTypeDirect ConversationType = "direct"
	ConversationTypeGroup  ConversationType = "group"
)

// Conversation represents a conversation (direct or group)
type Conversation struct {
	ID            uuid.UUID        `json:"id" db:"id"`
	Type          ConversationType `json:"type" db:"type"`
	Name          *string          `json:"name,omitempty" db:"name"`
	Description   *string          `json:"description,omitempty" db:"description"`
	AvatarURL     *string          `json:"avatar_url,omitempty" db:"avatar_url"`
	CreatedBy     *uuid.UUID       `json:"created_by,omitempty" db:"created_by"`
	CreatedAt     time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at" db:"updated_at"`
	LastMessageAt *time.Time       `json:"last_message_at,omitempty" db:"last_message_at"`
}

// MemberRole represents a member's role in a conversation
type MemberRole string

const (
	MemberRoleOwner  MemberRole = "owner"
	MemberRoleAdmin  MemberRole = "admin"
	MemberRoleMember MemberRole = "member"
)

// ConversationMember represents a user's membership in a conversation
type ConversationMember struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	ConversationID uuid.UUID  `json:"conversation_id" db:"conversation_id"`
	UserID         uuid.UUID  `json:"user_id" db:"user_id"`
	Role           MemberRole `json:"role" db:"role"`
	JoinedAt       time.Time  `json:"joined_at" db:"joined_at"`
	LeftAt         *time.Time `json:"left_at,omitempty" db:"left_at"`
	LastReadAt     *time.Time `json:"last_read_at,omitempty" db:"last_read_at"`
	IsMuted        bool       `json:"is_muted" db:"is_muted"`
}

// MessageType represents the type of message
type MessageType string

const (
	MessageTypeText   MessageType = "text"
	MessageTypeFile   MessageType = "file"
	MessageTypeImage  MessageType = "image"
	MessageTypeVideo  MessageType = "video"
	MessageTypeAudio  MessageType = "audio"
	MessageTypeSystem MessageType = "system"
)

// JSONB is a custom type for PostgreSQL JSONB
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// Message represents a message in a conversation
type Message struct {
	ID             uuid.UUID   `json:"id" db:"id"`
	ConversationID uuid.UUID   `json:"conversation_id" db:"conversation_id"`
	SenderID       uuid.UUID   `json:"sender_id" db:"sender_id"`
	Content        string      `json:"content" db:"content"`
	Type           MessageType `json:"type" db:"type"`
	Metadata       JSONB       `json:"metadata,omitempty" db:"metadata"`
	ReplyToID      *uuid.UUID  `json:"reply_to_id,omitempty" db:"reply_to_id"`
	IsEdited       bool        `json:"is_edited" db:"is_edited"`
	IsDeleted      bool        `json:"is_deleted" db:"is_deleted"`
	CreatedAt      time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at" db:"updated_at"`
	DeletedAt      *time.Time  `json:"deleted_at,omitempty" db:"deleted_at"`
}

// MessageReaction represents a reaction to a message
type MessageReaction struct {
	ID        uuid.UUID `json:"id" db:"id"`
	MessageID uuid.UUID `json:"message_id" db:"message_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Emoji     string    `json:"emoji" db:"emoji"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// MessageReadReceipt represents when a user read a message
type MessageReadReceipt struct {
	ID        uuid.UUID `json:"id" db:"id"`
	MessageID uuid.UUID `json:"message_id" db:"message_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	ReadAt    time.Time `json:"read_at" db:"read_at"`
}

// File represents a file attachment
type File struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	MessageID        *uuid.UUID `json:"message_id,omitempty" db:"message_id"`
	UploaderID       uuid.UUID  `json:"uploader_id" db:"uploader_id"`
	Filename         string     `json:"filename" db:"filename"`
	OriginalFilename string     `json:"original_filename" db:"original_filename"`
	MimeType         string     `json:"mime_type" db:"mime_type"`
	SizeBytes        int64      `json:"size_bytes" db:"size_bytes"`
	StoragePath      string     `json:"storage_path" db:"storage_path"`
	ThumbnailPath    *string    `json:"thumbnail_path,omitempty" db:"thumbnail_path"`
	Metadata         JSONB      `json:"metadata,omitempty" db:"metadata"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}

// TypingIndicator represents a user typing in a conversation
type TypingIndicator struct {
	ID             uuid.UUID `json:"id" db:"id"`
	ConversationID uuid.UUID `json:"conversation_id" db:"conversation_id"`
	UserID         uuid.UUID `json:"user_id" db:"user_id"`
	StartedAt      time.Time `json:"started_at" db:"started_at"`
	ExpiresAt      time.Time `json:"expires_at" db:"expires_at"`
}

// DTOs (Data Transfer Objects)

// CreateConversationRequest is the request to create a conversation
type CreateConversationRequest struct {
	Type        ConversationType `json:"type" binding:"required"`
	Name        *string          `json:"name,omitempty"`
	Description *string          `json:"description,omitempty"`
	MemberIDs   []uuid.UUID      `json:"member_ids" binding:"required"`
}

// UpdateConversationRequest is the request to update a conversation
type UpdateConversationRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
}

// SendMessageRequest is the request to send a message
type SendMessageRequest struct {
	Content   string     `json:"content" binding:"required"`
	Type      string     `json:"type,omitempty"`
	ReplyToID *uuid.UUID `json:"reply_to_id,omitempty"`
	Metadata  JSONB      `json:"metadata,omitempty"`
}

// EditMessageRequest is the request to edit a message
type EditMessageRequest struct {
	Content string `json:"content" binding:"required"`
}

// AddReactionRequest is the request to add a reaction
type AddReactionRequest struct {
	Emoji string `json:"emoji" binding:"required"`
}

// ConversationWithDetails includes additional conversation information
type ConversationWithDetails struct {
	Conversation
	Members      []User   `json:"members,omitempty"`
	UnreadCount  int      `json:"unread_count"`
	LastMessage  *Message `json:"last_message,omitempty"`
	MemberCount  int      `json:"member_count"`
}

// MessageWithDetails includes additional message information
type MessageWithDetails struct {
	Message
	Sender      *User               `json:"sender,omitempty"`
	Reactions   []MessageReaction   `json:"reactions,omitempty"`
	ReadBy      []MessageReadReceipt `json:"read_by,omitempty"`
	ReplyTo     *Message            `json:"reply_to,omitempty"`
	Attachments []File              `json:"attachments,omitempty"`
}

// WebSocket message types
type WSMessageType string

const (
	WSMessageTypeSend             WSMessageType = "message.send"
	WSMessageTypeNew              WSMessageType = "message.new"
	WSMessageTypeEdit             WSMessageType = "message.edit"
	WSMessageTypeEdited           WSMessageType = "message.edited"
	WSMessageTypeDelete           WSMessageType = "message.delete"
	WSMessageTypeDeleted          WSMessageType = "message.deleted"
	WSMessageTypeTypingStart      WSMessageType = "typing.start"
	WSMessageTypeTypingStop       WSMessageType = "typing.stop"
	WSMessageTypeTypingIndicator  WSMessageType = "typing.indicator"
	WSMessageTypePresenceUpdate   WSMessageType = "presence.update"
	WSMessageTypePresenceChanged  WSMessageType = "presence.changed"
	WSMessageTypeRead             WSMessageType = "message.read"
	WSMessageTypeReadReceipt      WSMessageType = "message.read_receipt"
	WSMessageTypeReactionAdded    WSMessageType = "reaction.added"
	WSMessageTypeReactionRemoved  WSMessageType = "reaction.removed"
	WSMessageTypeError            WSMessageType = "error"
)

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type WSMessageType `json:"type"`
	Data interface{}   `json:"data,omitempty"`
}

// WSTypingIndicator is sent when a user is typing
type WSTypingIndicator struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	UserID         uuid.UUID `json:"user_id"`
	IsTyping       bool      `json:"is_typing"`
}

// WSPresenceUpdate is sent when a user's presence changes
type WSPresenceUpdate struct {
	UserID uuid.UUID `json:"user_id"`
	Status string    `json:"status"`
}

// WSError represents an error sent over WebSocket
type WSError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
