package websocket

import (
	"context"
	"encoding/json"
	"time"

	"github.com/b25/services/messaging/internal/models"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024 // 512KB
)

// Client represents a WebSocket client connection
type Client struct {
	ID     uuid.UUID
	UserID uuid.UUID

	// The websocket connection
	Conn *websocket.Conn

	// Hub reference
	Hub *Hub

	// Buffered channel of outbound messages
	Send chan []byte

	// Ping channel
	Ping chan bool

	// Message handler
	MessageHandler MessageHandler

	// Logger
	Logger zerolog.Logger

	// Context
	ctx    context.Context
	cancel context.CancelFunc
}

// MessageHandler handles incoming WebSocket messages
type MessageHandler interface {
	HandleMessage(ctx context.Context, client *Client, msg models.WSMessage) error
}

// NewClient creates a new WebSocket client
func NewClient(userID uuid.UUID, conn *websocket.Conn, hub *Hub, handler MessageHandler, logger zerolog.Logger) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		ID:             uuid.New(),
		UserID:         userID,
		Conn:           conn,
		Hub:            hub,
		Send:           make(chan []byte, 256),
		Ping:           make(chan bool, 1),
		MessageHandler: handler,
		Logger:         logger,
		ctx:            ctx,
		cancel:         cancel,
	}
}

// ReadPump pumps messages from the websocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister(c)
		c.Conn.Close()
		c.cancel()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			_, message, err := c.Conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.Logger.Error().Err(err).Msg("WebSocket read error")
				}
				return
			}

			// Parse and handle message
			var wsMsg models.WSMessage
			if err := json.Unmarshal(message, &wsMsg); err != nil {
				c.Logger.Error().Err(err).Msg("Failed to unmarshal WebSocket message")
				c.sendError("INVALID_MESSAGE", "Invalid message format")
				continue
			}

			// Handle the message
			if err := c.MessageHandler.HandleMessage(c.ctx, c, wsMsg); err != nil {
				c.Logger.Error().Err(err).
					Str("message_type", string(wsMsg.Type)).
					Msg("Failed to handle WebSocket message")
				c.sendError("HANDLE_ERROR", err.Error())
			}
		}
	}
}

// WritePump pumps messages from the hub to the websocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
		c.cancel()
	}()

	for {
		select {
		case <-c.ctx.Done():
			return

		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-c.Ping:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// sendError sends an error message to the client
func (c *Client) sendError(code, message string) {
	errMsg := models.WSMessage{
		Type: models.WSMessageTypeError,
		Data: models.WSError{
			Code:    code,
			Message: message,
		},
	}

	data, err := json.Marshal(errMsg)
	if err != nil {
		c.Logger.Error().Err(err).Msg("Failed to marshal error message")
		return
	}

	select {
	case c.Send <- data:
	default:
		c.Logger.Warn().Msg("Failed to send error to client: buffer full")
	}
}

// Close closes the client connection
func (c *Client) Close() {
	c.cancel()
	c.Conn.Close()
}
