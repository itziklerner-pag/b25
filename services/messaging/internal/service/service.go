package service

import (
	"context"
	"fmt"
	"time"

	"github.com/b25/services/messaging/internal/models"
	"github.com/b25/services/messaging/internal/repository"
	"github.com/b25/services/messaging/internal/websocket"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// MessagingService provides business logic for the messaging service
type MessagingService struct {
	repo   repository.Repository
	hub    *websocket.Hub
	logger zerolog.Logger
}

// NewMessagingService creates a new messaging service
func NewMessagingService(repo repository.Repository, hub *websocket.Hub, logger zerolog.Logger) *MessagingService {
	return &MessagingService{
		repo:   repo,
		hub:    hub,
		logger: logger,
	}
}

// User operations

// CreateUser creates a new user
func (s *MessagingService) CreateUser(ctx context.Context, req *models.User) (*models.User, error) {
	if req.ID == uuid.Nil {
		req.ID = uuid.New()
	}
	if req.Status == "" {
		req.Status = "offline"
	}

	if err := s.repo.CreateUser(ctx, req); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return req, nil
}

// GetUser retrieves a user by ID
func (s *MessagingService) GetUser(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	return s.repo.GetUser(ctx, userID)
}

// UpdateUserStatus updates a user's online status
func (s *MessagingService) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) error {
	if err := s.repo.UpdateUserStatus(ctx, userID, status); err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	// Broadcast status change
	s.hub.Broadcast(models.WSMessageTypePresenceChanged, models.WSPresenceUpdate{
		UserID: userID,
		Status: status,
	})

	return nil
}

// Conversation operations

// CreateConversation creates a new conversation
func (s *MessagingService) CreateConversation(ctx context.Context, creatorID uuid.UUID, req *models.CreateConversationRequest) (*models.ConversationWithDetails, error) {
	// For direct conversations, check if one already exists
	if req.Type == models.ConversationTypeDirect {
		if len(req.MemberIDs) != 2 {
			return nil, fmt.Errorf("direct conversation must have exactly 2 members")
		}

		existingConv, err := s.repo.GetDirectConversation(ctx, req.MemberIDs[0], req.MemberIDs[1])
		if err == nil && existingConv != nil {
			return s.repo.GetConversationWithDetails(ctx, existingConv.ID, creatorID)
		}
	}

	// Create new conversation
	conv := &models.Conversation{
		ID:          uuid.New(),
		Type:        req.Type,
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   &creatorID,
	}

	if err := s.repo.CreateConversation(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	// Add members
	for _, memberID := range req.MemberIDs {
		role := models.MemberRoleMember
		if memberID == creatorID {
			role = models.MemberRoleOwner
		}

		member := &models.ConversationMember{
			ID:             uuid.New(),
			ConversationID: conv.ID,
			UserID:         memberID,
			Role:           role,
		}

		if err := s.repo.AddConversationMember(ctx, member); err != nil {
			return nil, fmt.Errorf("failed to add member: %w", err)
		}
	}

	return s.repo.GetConversationWithDetails(ctx, conv.ID, creatorID)
}

// GetConversation retrieves a conversation
func (s *MessagingService) GetConversation(ctx context.Context, conversationID, userID uuid.UUID) (*models.ConversationWithDetails, error) {
	// Check if user is a member
	isMember, err := s.repo.IsConversationMember(ctx, conversationID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, fmt.Errorf("user is not a member of this conversation")
	}

	return s.repo.GetConversationWithDetails(ctx, conversationID, userID)
}

// ListUserConversations lists all conversations for a user
func (s *MessagingService) ListUserConversations(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.ConversationWithDetails, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return s.repo.ListUserConversations(ctx, userID, limit, offset)
}

// UpdateConversation updates a conversation
func (s *MessagingService) UpdateConversation(ctx context.Context, conversationID, userID uuid.UUID, req *models.UpdateConversationRequest) (*models.ConversationWithDetails, error) {
	// Check if user is a member with appropriate permissions
	isMember, err := s.repo.IsConversationMember(ctx, conversationID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, fmt.Errorf("user is not a member of this conversation")
	}

	conv, err := s.repo.GetConversation(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		conv.Name = req.Name
	}
	if req.Description != nil {
		conv.Description = req.Description
	}
	if req.AvatarURL != nil {
		conv.AvatarURL = req.AvatarURL
	}

	if err := s.repo.UpdateConversation(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to update conversation: %w", err)
	}

	return s.repo.GetConversationWithDetails(ctx, conversationID, userID)
}

// AddConversationMember adds a member to a conversation
func (s *MessagingService) AddConversationMember(ctx context.Context, conversationID, adderID, newMemberID uuid.UUID) error {
	// Check if adder is a member
	isMember, err := s.repo.IsConversationMember(ctx, conversationID, adderID)
	if err != nil {
		return err
	}
	if !isMember {
		return fmt.Errorf("user is not a member of this conversation")
	}

	member := &models.ConversationMember{
		ID:             uuid.New(),
		ConversationID: conversationID,
		UserID:         newMemberID,
		Role:           models.MemberRoleMember,
	}

	return s.repo.AddConversationMember(ctx, member)
}

// RemoveConversationMember removes a member from a conversation
func (s *MessagingService) RemoveConversationMember(ctx context.Context, conversationID, removerID, memberID uuid.UUID) error {
	// Check if remover is a member
	isMember, err := s.repo.IsConversationMember(ctx, conversationID, removerID)
	if err != nil {
		return err
	}
	if !isMember {
		return fmt.Errorf("user is not a member of this conversation")
	}

	return s.repo.RemoveConversationMember(ctx, conversationID, memberID)
}

// Message operations

// SendMessage sends a new message
func (s *MessagingService) SendMessage(ctx context.Context, conversationID, senderID uuid.UUID, req *models.SendMessageRequest) (*models.MessageWithDetails, error) {
	// Check if sender is a member
	isMember, err := s.repo.IsConversationMember(ctx, conversationID, senderID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, fmt.Errorf("user is not a member of this conversation")
	}

	msgType := models.MessageTypeText
	if req.Type != "" {
		msgType = models.MessageType(req.Type)
	}

	message := &models.Message{
		ID:             uuid.New(),
		ConversationID: conversationID,
		SenderID:       senderID,
		Content:        req.Content,
		Type:           msgType,
		Metadata:       req.Metadata,
		ReplyToID:      req.ReplyToID,
	}

	if err := s.repo.CreateMessage(ctx, message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Get message with details
	msgWithDetails, err := s.repo.GetMessageWithDetails(ctx, message.ID)
	if err != nil {
		return nil, err
	}

	// Broadcast to conversation members
	members, err := s.repo.GetConversationMembers(ctx, conversationID)
	if err == nil {
		memberIDs := make([]uuid.UUID, len(members))
		for i, m := range members {
			memberIDs[i] = m.ID
		}
		s.hub.SendToConversation(memberIDs, models.WSMessageTypeNew, msgWithDetails)
	}

	return msgWithDetails, nil
}

// GetConversationMessages retrieves messages for a conversation
func (s *MessagingService) GetConversationMessages(ctx context.Context, conversationID, userID uuid.UUID, limit, offset int) ([]models.MessageWithDetails, error) {
	// Check if user is a member
	isMember, err := s.repo.IsConversationMember(ctx, conversationID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, fmt.Errorf("user is not a member of this conversation")
	}

	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	return s.repo.ListConversationMessages(ctx, conversationID, limit, offset)
}

// EditMessage edits a message
func (s *MessagingService) EditMessage(ctx context.Context, messageID, userID uuid.UUID, content string) (*models.MessageWithDetails, error) {
	message, err := s.repo.GetMessage(ctx, messageID)
	if err != nil {
		return nil, err
	}

	if message.SenderID != userID {
		return nil, fmt.Errorf("user is not the sender of this message")
	}

	message.Content = content
	if err := s.repo.UpdateMessage(ctx, message); err != nil {
		return nil, fmt.Errorf("failed to update message: %w", err)
	}

	msgWithDetails, err := s.repo.GetMessageWithDetails(ctx, messageID)
	if err != nil {
		return nil, err
	}

	// Broadcast edit to conversation members
	members, err := s.repo.GetConversationMembers(ctx, message.ConversationID)
	if err == nil {
		memberIDs := make([]uuid.UUID, len(members))
		for i, m := range members {
			memberIDs[i] = m.ID
		}
		s.hub.SendToConversation(memberIDs, models.WSMessageTypeEdited, msgWithDetails)
	}

	return msgWithDetails, nil
}

// DeleteMessage deletes a message
func (s *MessagingService) DeleteMessage(ctx context.Context, messageID, userID uuid.UUID) error {
	message, err := s.repo.GetMessage(ctx, messageID)
	if err != nil {
		return err
	}

	if message.SenderID != userID {
		return fmt.Errorf("user is not the sender of this message")
	}

	if err := s.repo.SoftDeleteMessage(ctx, messageID); err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	// Broadcast deletion to conversation members
	members, err := s.repo.GetConversationMembers(ctx, message.ConversationID)
	if err == nil {
		memberIDs := make([]uuid.UUID, len(members))
		for i, m := range members {
			memberIDs[i] = m.ID
		}
		s.hub.SendToConversation(memberIDs, models.WSMessageTypeDeleted, map[string]uuid.UUID{
			"message_id": messageID,
		})
	}

	return nil
}

// SearchMessages searches messages across all conversations
func (s *MessagingService) SearchMessages(ctx context.Context, userID uuid.UUID, query string, limit, offset int) ([]models.MessageWithDetails, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return s.repo.SearchMessages(ctx, userID, query, limit, offset)
}

// AddReaction adds a reaction to a message
func (s *MessagingService) AddReaction(ctx context.Context, messageID, userID uuid.UUID, emoji string) error {
	message, err := s.repo.GetMessage(ctx, messageID)
	if err != nil {
		return err
	}

	// Check if user is a member
	isMember, err := s.repo.IsConversationMember(ctx, message.ConversationID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return fmt.Errorf("user is not a member of this conversation")
	}

	reaction := &models.MessageReaction{
		ID:        uuid.New(),
		MessageID: messageID,
		UserID:    userID,
		Emoji:     emoji,
	}

	if err := s.repo.AddReaction(ctx, reaction); err != nil {
		return fmt.Errorf("failed to add reaction: %w", err)
	}

	// Broadcast reaction to conversation members
	members, err := s.repo.GetConversationMembers(ctx, message.ConversationID)
	if err == nil {
		memberIDs := make([]uuid.UUID, len(members))
		for i, m := range members {
			memberIDs[i] = m.ID
		}
		s.hub.SendToConversation(memberIDs, models.WSMessageTypeReactionAdded, reaction)
	}

	return nil
}

// RemoveReaction removes a reaction from a message
func (s *MessagingService) RemoveReaction(ctx context.Context, messageID, userID uuid.UUID, emoji string) error {
	message, err := s.repo.GetMessage(ctx, messageID)
	if err != nil {
		return err
	}

	if err := s.repo.RemoveReaction(ctx, messageID, userID, emoji); err != nil {
		return fmt.Errorf("failed to remove reaction: %w", err)
	}

	// Broadcast reaction removal
	members, err := s.repo.GetConversationMembers(ctx, message.ConversationID)
	if err == nil {
		memberIDs := make([]uuid.UUID, len(members))
		for i, m := range members {
			memberIDs[i] = m.ID
		}
		s.hub.SendToConversation(memberIDs, models.WSMessageTypeReactionRemoved, map[string]interface{}{
			"message_id": messageID,
			"user_id":    userID,
			"emoji":      emoji,
		})
	}

	return nil
}

// MarkMessageAsRead marks a message as read
func (s *MessagingService) MarkMessageAsRead(ctx context.Context, messageID, userID uuid.UUID) error {
	message, err := s.repo.GetMessage(ctx, messageID)
	if err != nil {
		return err
	}

	// Check if user is a member
	isMember, err := s.repo.IsConversationMember(ctx, message.ConversationID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return fmt.Errorf("user is not a member of this conversation")
	}

	receipt := &models.MessageReadReceipt{
		ID:        uuid.New(),
		MessageID: messageID,
		UserID:    userID,
	}

	if err := s.repo.MarkMessageAsRead(ctx, receipt); err != nil {
		return fmt.Errorf("failed to mark message as read: %w", err)
	}

	// Update member's last read timestamp
	if err := s.repo.UpdateMemberLastRead(ctx, message.ConversationID, userID, time.Now()); err != nil {
		s.logger.Error().Err(err).Msg("Failed to update member last read timestamp")
	}

	// Broadcast read receipt
	s.hub.SendToUser(message.SenderID, models.WSMessageTypeReadReceipt, map[string]uuid.UUID{
		"message_id": messageID,
		"user_id":    userID,
	})

	return nil
}

// SetTypingIndicator sets or updates a typing indicator
func (s *MessagingService) SetTypingIndicator(ctx context.Context, conversationID, userID uuid.UUID) error {
	// Check if user is a member
	isMember, err := s.repo.IsConversationMember(ctx, conversationID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return fmt.Errorf("user is not a member of this conversation")
	}

	indicator := &models.TypingIndicator{
		ID:             uuid.New(),
		ConversationID: conversationID,
		UserID:         userID,
		StartedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(5 * time.Second),
	}

	if err := s.repo.SetTypingIndicator(ctx, indicator); err != nil {
		return fmt.Errorf("failed to set typing indicator: %w", err)
	}

	// Broadcast typing indicator
	members, err := s.repo.GetConversationMembers(ctx, conversationID)
	if err == nil {
		memberIDs := make([]uuid.UUID, 0, len(members))
		for _, m := range members {
			if m.ID != userID {
				memberIDs = append(memberIDs, m.ID)
			}
		}
		s.hub.SendToConversation(memberIDs, models.WSMessageTypeTypingIndicator, models.WSTypingIndicator{
			ConversationID: conversationID,
			UserID:         userID,
			IsTyping:       true,
		})
	}

	return nil
}

// RemoveTypingIndicator removes a typing indicator
func (s *MessagingService) RemoveTypingIndicator(ctx context.Context, conversationID, userID uuid.UUID) error {
	if err := s.repo.RemoveTypingIndicator(ctx, conversationID, userID); err != nil {
		return fmt.Errorf("failed to remove typing indicator: %w", err)
	}

	// Broadcast typing stopped
	members, err := s.repo.GetConversationMembers(ctx, conversationID)
	if err == nil {
		memberIDs := make([]uuid.UUID, 0, len(members))
		for _, m := range members {
			if m.ID != userID {
				memberIDs = append(memberIDs, m.ID)
			}
		}
		s.hub.SendToConversation(memberIDs, models.WSMessageTypeTypingIndicator, models.WSTypingIndicator{
			ConversationID: conversationID,
			UserID:         userID,
			IsTyping:       false,
		})
	}

	return nil
}

// GetUnreadCount gets the unread message count for a conversation
func (s *MessagingService) GetUnreadCount(ctx context.Context, conversationID, userID uuid.UUID) (int, error) {
	return s.repo.GetUnreadCount(ctx, conversationID, userID)
}

// GetOnlineUsers returns list of currently online users
func (s *MessagingService) GetOnlineUsers() []uuid.UUID {
	return s.hub.GetConnectedUsers()
}

// IsUserOnline checks if a user is currently online
func (s *MessagingService) IsUserOnline(userID uuid.UUID) bool {
	return s.hub.IsUserConnected(userID)
}
