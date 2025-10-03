package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/yourusername/b25/services/dashboard-server/internal/aggregator"
	"github.com/yourusername/b25/services/dashboard-server/internal/broadcaster"
	"github.com/yourusername/b25/services/dashboard-server/internal/metrics"
	"github.com/yourusername/b25/services/dashboard-server/internal/types"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implement proper origin checking in production
		return true
	},
}

type Server struct {
	clients        map[string]*Client
	clientsMu      sync.RWMutex
	aggregator     *aggregator.Aggregator
	broadcaster    *broadcaster.Broadcaster
	logger         zerolog.Logger
	shutdownCtx    context.Context
	shutdownCancel context.CancelFunc
}

type Client struct {
	ID            string
	Type          types.ClientType
	Conn          *websocket.Conn
	Subscriptions map[string]bool
	SendChan      chan []byte
	LastUpdate    time.Time
	LastState     *types.State
	Context       context.Context
	Cancel        context.CancelFunc
	Format        types.SerializationFormat
}

func NewServer(logger zerolog.Logger, aggregator *aggregator.Aggregator, broadcaster *broadcaster.Broadcaster) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		clients:        make(map[string]*Client),
		aggregator:     aggregator,
		broadcaster:    broadcaster,
		logger:         logger,
		shutdownCtx:    ctx,
		shutdownCancel: cancel,
	}
}

func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to upgrade WebSocket connection")
		return
	}

	// Parse client type from query parameter
	clientType := types.ClientTypeTUI
	if r.URL.Query().Get("type") == "web" {
		clientType = types.ClientTypeWeb
	}

	// Parse serialization format
	format := types.FormatMessagePack
	if r.URL.Query().Get("format") == "json" {
		format = types.FormatJSON
	}

	client := s.createClient(conn, clientType, format)
	s.registerClient(client)

	s.logger.Info().
		Str("client_id", client.ID).
		Str("client_type", clientType.String()).
		Str("format", format.String()).
		Msg("Client connected")

	// Start goroutines for this client
	go s.clientReader(client)
	go s.clientWriter(client)

	// Register client with broadcaster
	s.broadcaster.RegisterClient(client.ID, client.Type, client.SendChan, format)
}

func (s *Server) HandleHistory(w http.ResponseWriter, r *http.Request) {
	// REST API for historical queries
	w.Header().Set("Content-Type", "application/json")

	// Get query parameters
	dataType := r.URL.Query().Get("type")
	symbol := r.URL.Query().Get("symbol")
	limit := r.URL.Query().Get("limit")

	if dataType == "" {
		http.Error(w, `{"error":"type parameter required"}`, http.StatusBadRequest)
		return
	}

	// TODO: Implement historical data retrieval from Redis/database
	response := map[string]interface{}{
		"type":   dataType,
		"symbol": symbol,
		"limit":  limit,
		"data":   []interface{}{},
	}

	json.NewEncoder(w).Encode(response)
}

func (s *Server) createClient(conn *websocket.Conn, clientType types.ClientType, format types.SerializationFormat) *Client {
	ctx, cancel := context.WithCancel(s.shutdownCtx)
	return &Client{
		ID:            generateClientID(),
		Type:          clientType,
		Conn:          conn,
		Subscriptions: make(map[string]bool),
		SendChan:      make(chan []byte, 256),
		LastUpdate:    time.Now(),
		Context:       ctx,
		Cancel:        cancel,
		Format:        format,
	}
}

func (s *Server) registerClient(client *Client) {
	s.clientsMu.Lock()
	s.clients[client.ID] = client
	s.clientsMu.Unlock()

	metrics.IncrementConnectedClients(client.Type.String())
}

func (s *Server) unregisterClient(client *Client) {
	s.clientsMu.Lock()
	delete(s.clients, client.ID)
	s.clientsMu.Unlock()

	client.Cancel()
	close(client.SendChan)

	s.logger.Info().
		Str("client_id", client.ID).
		Msg("Client disconnected")

	metrics.DecrementConnectedClients(client.Type.String())

	// Unregister from broadcaster
	s.broadcaster.UnregisterClient(client.ID)
}

func (s *Server) clientReader(client *Client) {
	defer s.unregisterClient(client)

	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		select {
		case <-client.Context.Done():
			return
		default:
			_, message, err := client.Conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					s.logger.Error().Err(err).Str("client_id", client.ID).Msg("WebSocket read error")
				}
				return
			}

			s.handleClientMessage(client, message)
		}
	}
}

func (s *Server) clientWriter(client *Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-client.Context.Done():
			return
		case message := <-client.SendChan:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
				s.logger.Error().Err(err).Str("client_id", client.ID).Msg("WebSocket write error")
				return
			}
		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (s *Server) handleClientMessage(client *Client, message []byte) {
	var msg types.ClientMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		s.logger.Error().Err(err).Str("client_id", client.ID).Msg("Failed to parse client message")
		return
	}

	metrics.RecordMessageReceived(msg.Type)

	switch msg.Type {
	case "subscribe":
		s.handleSubscribe(client, msg.Channels)
	case "unsubscribe":
		s.handleUnsubscribe(client, msg.Channels)
	case "refresh":
		s.handleRefresh(client)
	default:
		s.logger.Warn().Str("type", msg.Type).Msg("Unknown message type")
	}
}

func (s *Server) handleSubscribe(client *Client, channels []string) {
	for _, ch := range channels {
		client.Subscriptions[ch] = true
		metrics.IncrementClientSubscriptions(ch)
	}

	s.logger.Info().
		Str("client_id", client.ID).
		Strs("channels", channels).
		Msg("Client subscribed")

	// Update broadcaster subscriptions
	s.broadcaster.UpdateSubscriptions(client.ID, client.Subscriptions)

	// Send initial snapshot
	s.handleRefresh(client)
}

func (s *Server) handleUnsubscribe(client *Client, channels []string) {
	for _, ch := range channels {
		delete(client.Subscriptions, ch)
		metrics.DecrementClientSubscriptions(ch)
	}

	s.logger.Info().
		Str("client_id", client.ID).
		Strs("channels", channels).
		Msg("Client unsubscribed")

	// Update broadcaster subscriptions
	s.broadcaster.UpdateSubscriptions(client.ID, client.Subscriptions)
}

func (s *Server) handleRefresh(client *Client) {
	state := s.aggregator.GetFullState()
	filteredState := s.filterStateBySubscriptions(state, client.Subscriptions)
	s.sendFullState(client, filteredState)
}

func (s *Server) filterStateBySubscriptions(state *types.State, subscriptions map[string]bool) *types.State {
	filtered := &types.State{
		Timestamp: state.Timestamp,
		Sequence:  state.Sequence,
	}

	if subscriptions["market_data"] {
		filtered.MarketData = state.MarketData
	}
	if subscriptions["orders"] {
		filtered.Orders = state.Orders
	}
	if subscriptions["positions"] {
		filtered.Positions = state.Positions
	}
	if subscriptions["account"] {
		filtered.Account = state.Account
	}
	if subscriptions["strategies"] {
		filtered.Strategies = state.Strategies
	}

	return filtered
}

func (s *Server) sendFullState(client *Client, state *types.State) {
	msg := types.ServerMessage{
		Type:      "snapshot",
		Sequence:  state.Sequence,
		Timestamp: time.Now(),
		Data:      state,
	}

	message, err := s.serializeMessage(client.Format, msg)
	if err != nil {
		s.logger.Error().Err(err).Str("client_id", client.ID).Msg("Failed to serialize full state")
		return
	}

	select {
	case client.SendChan <- message:
		client.LastState = state
		metrics.RecordMessageSent(client.Type.String(), "snapshot")
		metrics.RecordMessageSize(client.Format.String(), "snapshot", len(message))
	default:
		s.logger.Warn().Str("client_id", client.ID).Msg("Client send buffer full")
	}
}

func (s *Server) serializeMessage(format types.SerializationFormat, msg types.ServerMessage) ([]byte, error) {
	start := time.Now()
	defer func() {
		metrics.RecordSerializationDuration(format.String(), time.Since(start).Seconds())
	}()

	switch format {
	case types.FormatMessagePack:
		return msgpack.Marshal(msg)
	case types.FormatJSON:
		return json.Marshal(msg)
	default:
		return nil, fmt.Errorf("unsupported serialization format: %v", format)
	}
}

func generateClientID() string {
	return fmt.Sprintf("client-%d", time.Now().UnixNano())
}
