package websocket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/b25/services/messaging/internal/models"
	"github.com/b25/services/messaging/internal/service"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Handler implements MessageHandler interface
type Handler struct {
	service *service.MessagingService
	logger  zerolog.Logger
}

// NewHandler creates a new WebSocket message handler
func NewHandler(svc *service.MessagingService, logger zerolog.Logger) *Handler {
	return &Handler{
		service: svc,
		logger:  logger,
	}
}

// HandleMessage handles incoming WebSocket messages
func (h *Handler) HandleMessage(ctx context.Context, client *Client, msg models.WSMessage) error {
	switch msg.Type {
	case models.WSMessageTypeSend:
		return h.handleSendMessage(ctx, client, msg.Data)
	case models.WSMessageTypeTypingStart:
		return h.handleTypingStart(ctx, client, msg.Data)
	case models.WSMessageTypeTypingStop:
		return h.handleTypingStop(ctx, client, msg.Data)
	case models.WSMessageTypePresenceUpdate:
		return h.handlePresenceUpdate(ctx, client, msg.Data)
	case models.WSMessageTypeRead:
		return h.handleMarkAsRead(ctx, client, msg.Data)
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

func (h *Handler) handleSendMessage(ctx context.Context, client *Client, data interface{}) error {
	var payload struct {
		ConversationID string                   `json:"conversation_id"`
		Content        string                   `json:"content"`
		Type           string                   `json:"type"`
		ReplyToID      *string                  `json:"reply_to_id"`
		Metadata       models.JSONB             `json:"metadata"`
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(dataBytes, &payload); err != nil {
		return err
	}

	conversationID, err := uuid.Parse(payload.ConversationID)
	if err != nil {
		return fmt.Errorf("invalid conversation_id")
	}

	req := &models.SendMessageRequest{
		Content:  payload.Content,
		Type:     payload.Type,
		Metadata: payload.Metadata,
	}

	if payload.ReplyToID != nil {
		replyTo, err := uuid.Parse(*payload.ReplyToID)
		if err == nil {
			req.ReplyToID = &replyTo
		}
	}

	_, err = h.service.SendMessage(ctx, conversationID, client.UserID, req)
	return err
}

func (h *Handler) handleTypingStart(ctx context.Context, client *Client, data interface{}) error {
	var payload struct {
		ConversationID string `json:"conversation_id"`
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(dataBytes, &payload); err != nil {
		return err
	}

	conversationID, err := uuid.Parse(payload.ConversationID)
	if err != nil {
		return fmt.Errorf("invalid conversation_id")
	}

	return h.service.SetTypingIndicator(ctx, conversationID, client.UserID)
}

func (h *Handler) handleTypingStop(ctx context.Context, client *Client, data interface{}) error {
	var payload struct {
		ConversationID string `json:"conversation_id"`
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(dataBytes, &payload); err != nil {
		return err
	}

	conversationID, err := uuid.Parse(payload.ConversationID)
	if err != nil {
		return fmt.Errorf("invalid conversation_id")
	}

	return h.service.RemoveTypingIndicator(ctx, conversationID, client.UserID)
}

func (h *Handler) handlePresenceUpdate(ctx context.Context, client *Client, data interface{}) error {
	var payload struct {
		Status string `json:"status"`
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(dataBytes, &payload); err != nil {
		return err
	}

	return h.service.UpdateUserStatus(ctx, client.UserID, payload.Status)
}

func (h *Handler) handleMarkAsRead(ctx context.Context, client *Client, data interface{}) error {
	var payload struct {
		MessageID string `json:"message_id"`
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(dataBytes, &payload); err != nil {
		return err
	}

	messageID, err := uuid.Parse(payload.MessageID)
	if err != nil {
		return fmt.Errorf("invalid message_id")
	}

	return h.service.MarkMessageAsRead(ctx, messageID, client.UserID)
}
