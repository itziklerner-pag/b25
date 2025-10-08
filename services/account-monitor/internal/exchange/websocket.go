package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/yourorg/b25/services/account-monitor/internal/config"
	"github.com/yourorg/b25/services/account-monitor/internal/metrics"
)

type WebSocketClient struct {
	config       config.ExchangeConfig
	binance      *BinanceClient
	conn         *websocket.Conn
	logger       *zap.Logger
	listenKey    string
	connected    bool
	reconnectCh  chan struct{}
}

type UserDataEvent struct {
	Type string
	Data interface{}
}

type EventHandler func(event UserDataEvent)

func NewWebSocketClient(cfg config.ExchangeConfig, logger *zap.Logger) *WebSocketClient {
	return &WebSocketClient{
		config:      cfg,
		binance:     NewBinanceClient(cfg, logger),
		logger:      logger,
		reconnectCh: make(chan struct{}, 1),
	}
}

// Start starts the WebSocket client
func (w *WebSocketClient) Start(ctx context.Context, handler EventHandler) error {
	for {
		if err := w.connect(ctx); err != nil {
			w.logger.Error("WebSocket connection failed", zap.Error(err))
			select {
			case <-time.After(w.config.WebSocket.ReconnectInterval):
				metrics.WebSocketReconnects.Inc()
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// Listen for messages
		go w.listen(ctx, handler)

		// Keep alive ticker
		keepAliveTicker := time.NewTicker(30 * time.Minute)
		defer keepAliveTicker.Stop()

		// Ping ticker
		pingTicker := time.NewTicker(w.config.WebSocket.PingInterval)
		defer pingTicker.Stop()

		for {
			select {
			case <-keepAliveTicker.C:
				if err := w.binance.KeepAliveListenKey(ctx, w.listenKey); err != nil {
					w.logger.Error("Failed to keep alive listen key", zap.Error(err))
				}

			case <-pingTicker.C:
				if err := w.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
					w.logger.Error("Failed to send ping", zap.Error(err))
					w.connected = false
					w.reconnectCh <- struct{}{}
				}

			case <-w.reconnectCh:
				w.logger.Info("Reconnecting WebSocket")
				w.conn.Close()
				metrics.WebSocketReconnects.Inc()
				break

			case <-ctx.Done():
				w.conn.Close()
				return ctx.Err()
			}

			if !w.connected {
				break
			}
		}
	}
}

// connect establishes WebSocket connection
func (w *WebSocketClient) connect(ctx context.Context) error {
	// Get listen key
	listenKey, err := w.binance.GetListenKey(ctx)
	if err != nil {
		return fmt.Errorf("failed to get listen key: %w", err)
	}
	w.listenKey = listenKey

	// Connect to Futures WebSocket user data stream
	wsURL := "wss://fstream.binance.com/ws/" + listenKey
	if w.config.Testnet {
		wsURL = "wss://stream.binancefuture.com/ws/" + listenKey
	}

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to dial Futures WebSocket: %w", err)
	}

	w.conn = conn
	w.connected = true
	w.logger.Info("Futures WebSocket connected", zap.String("url", wsURL))

	return nil
}

// listen listens for WebSocket messages
func (w *WebSocketClient) listen(ctx context.Context, handler EventHandler) {
	for {
		_, message, err := w.conn.ReadMessage()
		if err != nil {
			w.logger.Error("Failed to read WebSocket message", zap.Error(err))
			w.connected = false
			w.reconnectCh <- struct{}{}
			return
		}

		var rawEvent map[string]interface{}
		if err := json.Unmarshal(message, &rawEvent); err != nil {
			w.logger.Error("Failed to unmarshal WebSocket message", zap.Error(err))
			continue
		}

		eventType, ok := rawEvent["e"].(string)
		if !ok {
			w.logger.Warn("Unknown event format", zap.Any("event", rawEvent))
			continue
		}

		metrics.WebSocketMessagesReceived.WithLabelValues(eventType).Inc()

		event := UserDataEvent{
			Type: eventType,
			Data: rawEvent,
		}

		handler(event)
	}
}

// IsConnected returns connection status
func (w *WebSocketClient) IsConnected() bool {
	return w.connected
}

// Close closes the WebSocket connection
func (w *WebSocketClient) Close() error {
	if w.conn != nil {
		return w.conn.Close()
	}
	return nil
}
