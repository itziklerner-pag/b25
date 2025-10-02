package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/b25/services/messaging/internal/models"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// Registered clients
	clients map[uuid.UUID]*Client
	mu      sync.RWMutex

	// Inbound messages from clients
	broadcast chan *BroadcastMessage

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Logger
	logger zerolog.Logger

	// Context
	ctx    context.Context
	cancel context.CancelFunc
}

// BroadcastMessage represents a message to be broadcast to clients
type BroadcastMessage struct {
	Type       models.WSMessageType
	Data       interface{}
	TargetUser *uuid.UUID // If set, only send to this user
	ExceptUser *uuid.UUID // If set, don't send to this user
}

// NewHub creates a new Hub
func NewHub(logger zerolog.Logger) *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	return &Hub{
		clients:    make(map[uuid.UUID]*Client),
		broadcast:  make(chan *BroadcastMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Run starts the hub
func (h *Hub) Run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			h.logger.Info().Msg("Hub shutting down")
			return

		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.UserID] = client
			h.mu.Unlock()
			h.logger.Info().
				Str("user_id", client.UserID.String()).
				Str("client_id", client.ID.String()).
				Msg("Client registered")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				close(client.Send)
			}
			h.mu.Unlock()
			h.logger.Info().
				Str("user_id", client.UserID.String()).
				Str("client_id", client.ID.String()).
				Msg("Client unregistered")

		case message := <-h.broadcast:
			h.broadcastMessage(message)

		case <-ticker.C:
			// Ping all clients
			h.pingClients()
		}
	}
}

// broadcastMessage sends a message to all or specific clients
func (h *Hub) broadcastMessage(msg *BroadcastMessage) {
	wsMsg := models.WSMessage{
		Type: msg.Type,
		Data: msg.Data,
	}

	data, err := json.Marshal(wsMsg)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to marshal WebSocket message")
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for userID, client := range h.clients {
		// Skip if target user is set and doesn't match
		if msg.TargetUser != nil && userID != *msg.TargetUser {
			continue
		}

		// Skip if except user is set and matches
		if msg.ExceptUser != nil && userID == *msg.ExceptUser {
			continue
		}

		select {
		case client.Send <- data:
		default:
			// Client's send buffer is full, close the connection
			h.logger.Warn().
				Str("user_id", userID.String()).
				Msg("Client send buffer full, closing connection")
			go func(c *Client) {
				h.unregister <- c
			}(client)
		}
	}
}

// pingClients sends ping to all clients to keep connections alive
func (h *Hub) pingClients() {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		select {
		case client.Ping <- true:
		default:
			// Can't send ping, client might be stuck
		}
	}
}

// Broadcast sends a message to all connected clients
func (h *Hub) Broadcast(msgType models.WSMessageType, data interface{}) {
	h.broadcast <- &BroadcastMessage{
		Type: msgType,
		Data: data,
	}
}

// SendToUser sends a message to a specific user
func (h *Hub) SendToUser(userID uuid.UUID, msgType models.WSMessageType, data interface{}) {
	h.broadcast <- &BroadcastMessage{
		Type:       msgType,
		Data:       data,
		TargetUser: &userID,
	}
}

// SendToAllExcept sends a message to all users except the specified one
func (h *Hub) SendToAllExcept(exceptUserID uuid.UUID, msgType models.WSMessageType, data interface{}) {
	h.broadcast <- &BroadcastMessage{
		Type:       msgType,
		Data:       data,
		ExceptUser: &exceptUserID,
	}
}

// SendToConversation sends a message to all members of a conversation
func (h *Hub) SendToConversation(memberIDs []uuid.UUID, msgType models.WSMessageType, data interface{}) {
	wsMsg := models.WSMessage{
		Type: msgType,
		Data: data,
	}

	msgData, err := json.Marshal(wsMsg)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to marshal WebSocket message")
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, memberID := range memberIDs {
		if client, ok := h.clients[memberID]; ok {
			select {
			case client.Send <- msgData:
			default:
				h.logger.Warn().
					Str("user_id", memberID.String()).
					Msg("Failed to send to client")
			}
		}
	}
}

// Register registers a new client
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister unregisters a client
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Shutdown gracefully shuts down the hub
func (h *Hub) Shutdown() {
	h.cancel()
}

// GetConnectedUsers returns a list of currently connected user IDs
func (h *Hub) GetConnectedUsers() []uuid.UUID {
	h.mu.RLock()
	defer h.mu.RUnlock()

	userIDs := make([]uuid.UUID, 0, len(h.clients))
	for userID := range h.clients {
		userIDs = append(userIDs, userID)
	}
	return userIDs
}

// IsUserConnected checks if a user is currently connected
func (h *Hub) IsUserConnected(userID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	_, ok := h.clients[userID]
	return ok
}

// GetClientCount returns the number of connected clients
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.clients)
}
