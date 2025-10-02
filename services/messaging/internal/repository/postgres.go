package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/b25/services/messaging/internal/models"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// PostgresRepository implements Repository interface using PostgreSQL
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// User operations

func (r *PostgresRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, username, email, display_name, avatar_url, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		user.ID, user.Username, user.Email, user.DisplayName, user.AvatarURL, user.Status,
	).Scan(&user.CreatedAt, &user.UpdatedAt)
}

func (r *PostgresRepository) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, username, email, display_name, avatar_url, status, last_seen, created_at, updated_at
		FROM users WHERE id = $1
	`
	var user models.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.DisplayName, &user.AvatarURL,
		&user.Status, &user.LastSeen, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return &user, err
}

func (r *PostgresRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT id, username, email, display_name, avatar_url, status, last_seen, created_at, updated_at
		FROM users WHERE username = $1
	`
	var user models.User
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.DisplayName, &user.AvatarURL,
		&user.Status, &user.LastSeen, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return &user, err
}

func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, username, email, display_name, avatar_url, status, last_seen, created_at, updated_at
		FROM users WHERE email = $1
	`
	var user models.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.DisplayName, &user.AvatarURL,
		&user.Status, &user.LastSeen, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return &user, err
}

func (r *PostgresRepository) UpdateUser(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET display_name = $2, avatar_url = $3, status = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		user.ID, user.DisplayName, user.AvatarURL, user.Status,
	).Scan(&user.UpdatedAt)
}

func (r *PostgresRepository) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) error {
	query := `UPDATE users SET status = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, userID, status)
	return err
}

func (r *PostgresRepository) UpdateUserLastSeen(ctx context.Context, userID uuid.UUID, lastSeen time.Time) error {
	query := `UPDATE users SET last_seen = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, userID, lastSeen)
	return err
}

// Conversation operations

func (r *PostgresRepository) CreateConversation(ctx context.Context, conversation *models.Conversation) error {
	query := `
		INSERT INTO conversations (id, type, name, description, avatar_url, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		conversation.ID, conversation.Type, conversation.Name, conversation.Description,
		conversation.AvatarURL, conversation.CreatedBy,
	).Scan(&conversation.CreatedAt, &conversation.UpdatedAt)
}

func (r *PostgresRepository) GetConversation(ctx context.Context, id uuid.UUID) (*models.Conversation, error) {
	query := `
		SELECT id, type, name, description, avatar_url, created_by, created_at, updated_at, last_message_at
		FROM conversations WHERE id = $1
	`
	var conv models.Conversation
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&conv.ID, &conv.Type, &conv.Name, &conv.Description, &conv.AvatarURL,
		&conv.CreatedBy, &conv.CreatedAt, &conv.UpdatedAt, &conv.LastMessageAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("conversation not found")
	}
	return &conv, err
}

func (r *PostgresRepository) GetConversationWithDetails(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.ConversationWithDetails, error) {
	conv, err := r.GetConversation(ctx, id)
	if err != nil {
		return nil, err
	}

	details := &models.ConversationWithDetails{
		Conversation: *conv,
	}

	// Get members
	members, err := r.GetConversationMembers(ctx, id)
	if err == nil {
		details.Members = members
		details.MemberCount = len(members)
	}

	// Get unread count
	unreadCount, err := r.GetUnreadCount(ctx, id, userID)
	if err == nil {
		details.UnreadCount = unreadCount
	}

	// Get last message
	messages, err := r.ListConversationMessages(ctx, id, 1, 0)
	if err == nil && len(messages) > 0 {
		details.LastMessage = &messages[0].Message
	}

	return details, nil
}

func (r *PostgresRepository) ListUserConversations(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.ConversationWithDetails, error) {
	query := `
		SELECT DISTINCT c.id, c.type, c.name, c.description, c.avatar_url, c.created_by,
		       c.created_at, c.updated_at, c.last_message_at
		FROM conversations c
		INNER JOIN conversation_members cm ON cm.conversation_id = c.id
		WHERE cm.user_id = $1 AND cm.left_at IS NULL
		ORDER BY c.last_message_at DESC NULLS LAST, c.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []models.ConversationWithDetails
	for rows.Next() {
		var conv models.Conversation
		if err := rows.Scan(
			&conv.ID, &conv.Type, &conv.Name, &conv.Description, &conv.AvatarURL,
			&conv.CreatedBy, &conv.CreatedAt, &conv.UpdatedAt, &conv.LastMessageAt,
		); err != nil {
			return nil, err
		}

		details := models.ConversationWithDetails{
			Conversation: conv,
		}

		// Get additional details
		members, _ := r.GetConversationMembers(ctx, conv.ID)
		details.Members = members
		details.MemberCount = len(members)

		unreadCount, _ := r.GetUnreadCount(ctx, conv.ID, userID)
		details.UnreadCount = unreadCount

		conversations = append(conversations, details)
	}

	return conversations, rows.Err()
}

func (r *PostgresRepository) UpdateConversation(ctx context.Context, conversation *models.Conversation) error {
	query := `
		UPDATE conversations
		SET name = $2, description = $3, avatar_url = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		conversation.ID, conversation.Name, conversation.Description, conversation.AvatarURL,
	).Scan(&conversation.UpdatedAt)
}

func (r *PostgresRepository) DeleteConversation(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM conversations WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *PostgresRepository) GetDirectConversation(ctx context.Context, user1ID, user2ID uuid.UUID) (*models.Conversation, error) {
	query := `
		SELECT c.id, c.type, c.name, c.description, c.avatar_url, c.created_by,
		       c.created_at, c.updated_at, c.last_message_at
		FROM conversations c
		WHERE c.type = 'direct'
		AND c.id IN (
			SELECT cm1.conversation_id
			FROM conversation_members cm1
			INNER JOIN conversation_members cm2 ON cm1.conversation_id = cm2.conversation_id
			WHERE cm1.user_id = $1 AND cm2.user_id = $2
			AND cm1.left_at IS NULL AND cm2.left_at IS NULL
		)
		LIMIT 1
	`
	var conv models.Conversation
	err := r.db.QueryRowContext(ctx, query, user1ID, user2ID).Scan(
		&conv.ID, &conv.Type, &conv.Name, &conv.Description, &conv.AvatarURL,
		&conv.CreatedBy, &conv.CreatedAt, &conv.UpdatedAt, &conv.LastMessageAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &conv, err
}

// Conversation member operations

func (r *PostgresRepository) AddConversationMember(ctx context.Context, member *models.ConversationMember) error {
	query := `
		INSERT INTO conversation_members (id, conversation_id, user_id, role)
		VALUES ($1, $2, $3, $4)
		RETURNING joined_at
	`
	return r.db.QueryRowContext(ctx, query,
		member.ID, member.ConversationID, member.UserID, member.Role,
	).Scan(&member.JoinedAt)
}

func (r *PostgresRepository) RemoveConversationMember(ctx context.Context, conversationID, userID uuid.UUID) error {
	query := `
		UPDATE conversation_members
		SET left_at = CURRENT_TIMESTAMP
		WHERE conversation_id = $1 AND user_id = $2
	`
	_, err := r.db.ExecContext(ctx, query, conversationID, userID)
	return err
}

func (r *PostgresRepository) GetConversationMembers(ctx context.Context, conversationID uuid.UUID) ([]models.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.display_name, u.avatar_url, u.status,
		       u.last_seen, u.created_at, u.updated_at
		FROM users u
		INNER JOIN conversation_members cm ON cm.user_id = u.id
		WHERE cm.conversation_id = $1 AND cm.left_at IS NULL
		ORDER BY u.username
	`
	rows, err := r.db.QueryContext(ctx, query, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.DisplayName, &user.AvatarURL,
			&user.Status, &user.LastSeen, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

func (r *PostgresRepository) IsConversationMember(ctx context.Context, conversationID, userID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM conversation_members
			WHERE conversation_id = $1 AND user_id = $2 AND left_at IS NULL
		)
	`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, conversationID, userID).Scan(&exists)
	return exists, err
}

func (r *PostgresRepository) UpdateMemberLastRead(ctx context.Context, conversationID, userID uuid.UUID, lastReadAt time.Time) error {
	query := `
		UPDATE conversation_members
		SET last_read_at = $3
		WHERE conversation_id = $1 AND user_id = $2
	`
	_, err := r.db.ExecContext(ctx, query, conversationID, userID, lastReadAt)
	return err
}

func (r *PostgresRepository) MuteConversation(ctx context.Context, conversationID, userID uuid.UUID, muted bool) error {
	query := `
		UPDATE conversation_members
		SET is_muted = $3
		WHERE conversation_id = $1 AND user_id = $2
	`
	_, err := r.db.ExecContext(ctx, query, conversationID, userID, muted)
	return err
}

// Message operations

func (r *PostgresRepository) CreateMessage(ctx context.Context, message *models.Message) error {
	query := `
		INSERT INTO messages (id, conversation_id, sender_id, content, type, metadata, reply_to_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		message.ID, message.ConversationID, message.SenderID, message.Content,
		message.Type, message.Metadata, message.ReplyToID,
	).Scan(&message.CreatedAt, &message.UpdatedAt)
}

func (r *PostgresRepository) GetMessage(ctx context.Context, id uuid.UUID) (*models.Message, error) {
	query := `
		SELECT id, conversation_id, sender_id, content, type, metadata, reply_to_id,
		       is_edited, is_deleted, created_at, updated_at, deleted_at
		FROM messages WHERE id = $1
	`
	var msg models.Message
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Content, &msg.Type,
		&msg.Metadata, &msg.ReplyToID, &msg.IsEdited, &msg.IsDeleted,
		&msg.CreatedAt, &msg.UpdatedAt, &msg.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("message not found")
	}
	return &msg, err
}

func (r *PostgresRepository) GetMessageWithDetails(ctx context.Context, id uuid.UUID) (*models.MessageWithDetails, error) {
	msg, err := r.GetMessage(ctx, id)
	if err != nil {
		return nil, err
	}

	details := &models.MessageWithDetails{
		Message: *msg,
	}

	// Get sender
	sender, err := r.GetUser(ctx, msg.SenderID)
	if err == nil {
		details.Sender = sender
	}

	// Get reactions
	reactions, err := r.GetMessageReactions(ctx, id)
	if err == nil {
		details.Reactions = reactions
	}

	// Get read receipts
	readReceipts, err := r.GetMessageReadReceipts(ctx, id)
	if err == nil {
		details.ReadBy = readReceipts
	}

	// Get attachments
	files, err := r.GetMessageFiles(ctx, id)
	if err == nil {
		details.Attachments = files
	}

	return details, nil
}

func (r *PostgresRepository) ListConversationMessages(ctx context.Context, conversationID uuid.UUID, limit, offset int) ([]models.MessageWithDetails, error) {
	query := `
		SELECT id, conversation_id, sender_id, content, type, metadata, reply_to_id,
		       is_edited, is_deleted, created_at, updated_at, deleted_at
		FROM messages
		WHERE conversation_id = $1 AND is_deleted = FALSE
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, conversationID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.MessageWithDetails
	for rows.Next() {
		var msg models.Message
		if err := rows.Scan(
			&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Content, &msg.Type,
			&msg.Metadata, &msg.ReplyToID, &msg.IsEdited, &msg.IsDeleted,
			&msg.CreatedAt, &msg.UpdatedAt, &msg.DeletedAt,
		); err != nil {
			return nil, err
		}

		details := models.MessageWithDetails{Message: msg}

		// Get sender
		sender, _ := r.GetUser(ctx, msg.SenderID)
		details.Sender = sender

		messages = append(messages, details)
	}

	return messages, rows.Err()
}

func (r *PostgresRepository) UpdateMessage(ctx context.Context, message *models.Message) error {
	query := `
		UPDATE messages
		SET content = $2, is_edited = TRUE, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at
	`
	return r.db.QueryRowContext(ctx, query, message.ID, message.Content).Scan(&message.UpdatedAt)
}

func (r *PostgresRepository) DeleteMessage(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM messages WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *PostgresRepository) SoftDeleteMessage(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE messages
		SET is_deleted = TRUE, deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *PostgresRepository) SearchMessages(ctx context.Context, userID uuid.UUID, query string, limit, offset int) ([]models.MessageWithDetails, error) {
	sqlQuery := `
		SELECT DISTINCT m.id, m.conversation_id, m.sender_id, m.content, m.type, m.metadata,
		       m.reply_to_id, m.is_edited, m.is_deleted, m.created_at, m.updated_at, m.deleted_at
		FROM messages m
		INNER JOIN conversation_members cm ON cm.conversation_id = m.conversation_id
		WHERE cm.user_id = $1 AND cm.left_at IS NULL
		AND m.is_deleted = FALSE
		AND m.content ILIKE '%' || $2 || '%'
		ORDER BY m.created_at DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.QueryContext(ctx, sqlQuery, userID, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.MessageWithDetails
	for rows.Next() {
		var msg models.Message
		if err := rows.Scan(
			&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Content, &msg.Type,
			&msg.Metadata, &msg.ReplyToID, &msg.IsEdited, &msg.IsDeleted,
			&msg.CreatedAt, &msg.UpdatedAt, &msg.DeletedAt,
		); err != nil {
			return nil, err
		}

		details := models.MessageWithDetails{Message: msg}
		sender, _ := r.GetUser(ctx, msg.SenderID)
		details.Sender = sender

		messages = append(messages, details)
	}

	return messages, rows.Err()
}

// Message reaction operations

func (r *PostgresRepository) AddReaction(ctx context.Context, reaction *models.MessageReaction) error {
	query := `
		INSERT INTO message_reactions (id, message_id, user_id, emoji)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (message_id, user_id, emoji) DO NOTHING
		RETURNING created_at
	`
	return r.db.QueryRowContext(ctx, query,
		reaction.ID, reaction.MessageID, reaction.UserID, reaction.Emoji,
	).Scan(&reaction.CreatedAt)
}

func (r *PostgresRepository) RemoveReaction(ctx context.Context, messageID, userID uuid.UUID, emoji string) error {
	query := `DELETE FROM message_reactions WHERE message_id = $1 AND user_id = $2 AND emoji = $3`
	_, err := r.db.ExecContext(ctx, query, messageID, userID, emoji)
	return err
}

func (r *PostgresRepository) GetMessageReactions(ctx context.Context, messageID uuid.UUID) ([]models.MessageReaction, error) {
	query := `
		SELECT id, message_id, user_id, emoji, created_at
		FROM message_reactions WHERE message_id = $1
		ORDER BY created_at
	`
	rows, err := r.db.QueryContext(ctx, query, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reactions []models.MessageReaction
	for rows.Next() {
		var reaction models.MessageReaction
		if err := rows.Scan(
			&reaction.ID, &reaction.MessageID, &reaction.UserID, &reaction.Emoji, &reaction.CreatedAt,
		); err != nil {
			return nil, err
		}
		reactions = append(reactions, reaction)
	}

	return reactions, rows.Err()
}

// Message read receipt operations

func (r *PostgresRepository) MarkMessageAsRead(ctx context.Context, receipt *models.MessageReadReceipt) error {
	query := `
		INSERT INTO message_read_receipts (id, message_id, user_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (message_id, user_id) DO NOTHING
		RETURNING read_at
	`
	return r.db.QueryRowContext(ctx, query,
		receipt.ID, receipt.MessageID, receipt.UserID,
	).Scan(&receipt.ReadAt)
}

func (r *PostgresRepository) GetMessageReadReceipts(ctx context.Context, messageID uuid.UUID) ([]models.MessageReadReceipt, error) {
	query := `
		SELECT id, message_id, user_id, read_at
		FROM message_read_receipts WHERE message_id = $1
		ORDER BY read_at
	`
	rows, err := r.db.QueryContext(ctx, query, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var receipts []models.MessageReadReceipt
	for rows.Next() {
		var receipt models.MessageReadReceipt
		if err := rows.Scan(
			&receipt.ID, &receipt.MessageID, &receipt.UserID, &receipt.ReadAt,
		); err != nil {
			return nil, err
		}
		receipts = append(receipts, receipt)
	}

	return receipts, rows.Err()
}

func (r *PostgresRepository) GetUnreadCount(ctx context.Context, conversationID, userID uuid.UUID) (int, error) {
	query := `
		SELECT COALESCE(unread_count, 0)
		FROM unread_message_counts
		WHERE conversation_id = $1 AND user_id = $2
	`
	var count int
	err := r.db.QueryRowContext(ctx, query, conversationID, userID).Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return count, err
}

// File operations

func (r *PostgresRepository) CreateFile(ctx context.Context, file *models.File) error {
	query := `
		INSERT INTO files (id, message_id, uploader_id, filename, original_filename, mime_type,
		                   size_bytes, storage_path, thumbnail_path, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at
	`
	return r.db.QueryRowContext(ctx, query,
		file.ID, file.MessageID, file.UploaderID, file.Filename, file.OriginalFilename,
		file.MimeType, file.SizeBytes, file.StoragePath, file.ThumbnailPath, file.Metadata,
	).Scan(&file.CreatedAt)
}

func (r *PostgresRepository) GetFile(ctx context.Context, id uuid.UUID) (*models.File, error) {
	query := `
		SELECT id, message_id, uploader_id, filename, original_filename, mime_type,
		       size_bytes, storage_path, thumbnail_path, metadata, created_at
		FROM files WHERE id = $1
	`
	var file models.File
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&file.ID, &file.MessageID, &file.UploaderID, &file.Filename, &file.OriginalFilename,
		&file.MimeType, &file.SizeBytes, &file.StoragePath, &file.ThumbnailPath,
		&file.Metadata, &file.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("file not found")
	}
	return &file, err
}

func (r *PostgresRepository) GetMessageFiles(ctx context.Context, messageID uuid.UUID) ([]models.File, error) {
	query := `
		SELECT id, message_id, uploader_id, filename, original_filename, mime_type,
		       size_bytes, storage_path, thumbnail_path, metadata, created_at
		FROM files WHERE message_id = $1
		ORDER BY created_at
	`
	rows, err := r.db.QueryContext(ctx, query, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []models.File
	for rows.Next() {
		var file models.File
		if err := rows.Scan(
			&file.ID, &file.MessageID, &file.UploaderID, &file.Filename, &file.OriginalFilename,
			&file.MimeType, &file.SizeBytes, &file.StoragePath, &file.ThumbnailPath,
			&file.Metadata, &file.CreatedAt,
		); err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	return files, rows.Err()
}

func (r *PostgresRepository) DeleteFile(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM files WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// Typing indicator operations

func (r *PostgresRepository) SetTypingIndicator(ctx context.Context, indicator *models.TypingIndicator) error {
	query := `
		INSERT INTO typing_indicators (id, conversation_id, user_id, started_at, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (conversation_id, user_id)
		DO UPDATE SET started_at = $4, expires_at = $5
	`
	_, err := r.db.ExecContext(ctx, query,
		indicator.ID, indicator.ConversationID, indicator.UserID,
		indicator.StartedAt, indicator.ExpiresAt,
	)
	return err
}

func (r *PostgresRepository) GetTypingIndicators(ctx context.Context, conversationID uuid.UUID) ([]models.TypingIndicator, error) {
	query := `
		SELECT id, conversation_id, user_id, started_at, expires_at
		FROM typing_indicators
		WHERE conversation_id = $1 AND expires_at > CURRENT_TIMESTAMP
		ORDER BY started_at
	`
	rows, err := r.db.QueryContext(ctx, query, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indicators []models.TypingIndicator
	for rows.Next() {
		var indicator models.TypingIndicator
		if err := rows.Scan(
			&indicator.ID, &indicator.ConversationID, &indicator.UserID,
			&indicator.StartedAt, &indicator.ExpiresAt,
		); err != nil {
			return nil, err
		}
		indicators = append(indicators, indicator)
	}

	return indicators, rows.Err()
}

func (r *PostgresRepository) RemoveTypingIndicator(ctx context.Context, conversationID, userID uuid.UUID) error {
	query := `DELETE FROM typing_indicators WHERE conversation_id = $1 AND user_id = $2`
	_, err := r.db.ExecContext(ctx, query, conversationID, userID)
	return err
}

func (r *PostgresRepository) CleanupExpiredTypingIndicators(ctx context.Context) error {
	query := `DELETE FROM typing_indicators WHERE expires_at < CURRENT_TIMESTAMP`
	_, err := r.db.ExecContext(ctx, query)
	return err
}

// Transaction helpers

func (r *PostgresRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

func (r *PostgresRepository) CommitTx(tx *sql.Tx) error {
	return tx.Commit()
}

func (r *PostgresRepository) RollbackTx(tx *sql.Tx) error {
	return tx.Rollback()
}
