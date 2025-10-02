package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/yourorg/b25/services/search/internal/config"
	"github.com/yourorg/b25/services/search/internal/search"
	"github.com/yourorg/b25/services/search/pkg/models"
)

// Indexer handles real-time indexing of documents
type Indexer struct {
	es        *search.ElasticsearchClient
	natsConn  *nats.Conn
	config    *config.IndexerConfig
	subjects  *config.SubjectsConfig
	logger    *zap.Logger
	queue     chan *models.IndexRequest
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewIndexer creates a new indexer
func NewIndexer(
	es *search.ElasticsearchClient,
	natsConn *nats.Conn,
	cfg *config.IndexerConfig,
	subjects *config.SubjectsConfig,
	logger *zap.Logger,
) *Indexer {
	ctx, cancel := context.WithCancel(context.Background())

	return &Indexer{
		es:       es,
		natsConn: natsConn,
		config:   cfg,
		subjects: subjects,
		logger:   logger,
		queue:    make(chan *models.IndexRequest, cfg.QueueBuffer),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start starts the indexer workers and NATS subscribers
func (i *Indexer) Start() error {
	// Start worker pool
	for w := 0; w < i.config.Workers; w++ {
		i.wg.Add(1)
		go i.worker(w)
	}

	// Start flush ticker
	i.wg.Add(1)
	go i.flushTicker()

	// Subscribe to NATS subjects
	if err := i.subscribeToSubjects(); err != nil {
		i.Stop()
		return fmt.Errorf("failed to subscribe to subjects: %w", err)
	}

	i.logger.Info("Indexer started", zap.Int("workers", i.config.Workers))
	return nil
}

// Stop stops the indexer
func (i *Indexer) Stop() {
	i.logger.Info("Stopping indexer...")
	i.cancel()
	close(i.queue)
	i.wg.Wait()
	i.logger.Info("Indexer stopped")
}

// Index queues a document for indexing
func (i *Indexer) Index(req *models.IndexRequest) error {
	select {
	case i.queue <- req:
		return nil
	case <-i.ctx.Done():
		return fmt.Errorf("indexer is shutting down")
	default:
		return fmt.Errorf("indexer queue is full")
	}
}

// worker processes documents from the queue
func (i *Indexer) worker(id int) {
	defer i.wg.Done()

	batch := make([]*models.IndexRequest, 0, i.config.BatchSize)
	ticker := time.NewTicker(i.config.FlushInterval)
	defer ticker.Stop()

	i.logger.Debug("Worker started", zap.Int("worker_id", id))

	for {
		select {
		case req, ok := <-i.queue:
			if !ok {
				// Queue closed, flush remaining batch
				if len(batch) > 0 {
					i.flushBatch(batch)
				}
				return
			}

			batch = append(batch, req)

			// Flush if batch is full
			if len(batch) >= i.config.BatchSize {
				i.flushBatch(batch)
				batch = make([]*models.IndexRequest, 0, i.config.BatchSize)
			}

		case <-ticker.C:
			// Periodic flush
			if len(batch) > 0 {
				i.flushBatch(batch)
				batch = make([]*models.IndexRequest, 0, i.config.BatchSize)
			}

		case <-i.ctx.Done():
			// Shutdown, flush remaining batch
			if len(batch) > 0 {
				i.flushBatch(batch)
			}
			return
		}
	}
}

// flushTicker ensures periodic flushing even with low volume
func (i *Indexer) flushTicker() {
	defer i.wg.Done()

	ticker := time.NewTicker(i.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Workers handle their own flushing
		case <-i.ctx.Done():
			return
		}
	}
}

// flushBatch indexes a batch of documents
func (i *Indexer) flushBatch(batch []*models.IndexRequest) {
	if len(batch) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	bulkReq := &models.BulkIndexRequest{
		Documents: batch,
	}

	var resp *models.BulkIndexResponse
	var err error

	// Retry logic
	for attempt := 0; attempt <= i.config.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := i.config.RetryDelay * time.Duration(attempt)
			i.logger.Warn("Retrying bulk index",
				zap.Int("attempt", attempt),
				zap.Duration("delay", delay),
			)
			time.Sleep(delay)
		}

		resp, err = i.es.BulkIndex(ctx, bulkReq)
		if err == nil && resp.Success {
			break
		}
	}

	if err != nil {
		i.logger.Error("Failed to index batch after retries",
			zap.Error(err),
			zap.Int("batch_size", len(batch)),
		)
		return
	}

	if !resp.Success {
		i.logger.Error("Bulk index completed with errors",
			zap.Int("indexed", resp.Indexed),
			zap.Int("failed", resp.Failed),
		)
	} else {
		i.logger.Debug("Batch indexed successfully",
			zap.Int("count", resp.Indexed),
			zap.Int64("took_ms", resp.Took),
		)
	}
}

// subscribeToSubjects subscribes to NATS subjects for real-time updates
func (i *Indexer) subscribeToSubjects() error {
	// Subscribe to trades
	if _, err := i.natsConn.Subscribe(i.subjects.Trades, i.handleTradeMessage); err != nil {
		return fmt.Errorf("failed to subscribe to trades: %w", err)
	}

	// Subscribe to orders
	if _, err := i.natsConn.Subscribe(i.subjects.Orders, i.handleOrderMessage); err != nil {
		return fmt.Errorf("failed to subscribe to orders: %w", err)
	}

	// Subscribe to strategies
	if _, err := i.natsConn.Subscribe(i.subjects.Strategies, i.handleStrategyMessage); err != nil {
		return fmt.Errorf("failed to subscribe to strategies: %w", err)
	}

	// Subscribe to market data
	if _, err := i.natsConn.Subscribe(i.subjects.MarketData, i.handleMarketDataMessage); err != nil {
		return fmt.Errorf("failed to subscribe to market data: %w", err)
	}

	// Subscribe to logs
	if _, err := i.natsConn.Subscribe(i.subjects.Logs, i.handleLogMessage); err != nil {
		return fmt.Errorf("failed to subscribe to logs: %w", err)
	}

	i.logger.Info("Subscribed to NATS subjects",
		zap.String("trades", i.subjects.Trades),
		zap.String("orders", i.subjects.Orders),
		zap.String("strategies", i.subjects.Strategies),
		zap.String("market_data", i.subjects.MarketData),
		zap.String("logs", i.subjects.Logs),
	)

	return nil
}

// Message handlers

func (i *Indexer) handleTradeMessage(msg *nats.Msg) {
	var trade models.Trade
	if err := json.Unmarshal(msg.Data, &trade); err != nil {
		i.logger.Error("Failed to unmarshal trade message", zap.Error(err))
		return
	}

	doc := map[string]interface{}{
		"id":              trade.ID,
		"symbol":          trade.Symbol,
		"side":            trade.Side,
		"type":            trade.Type,
		"quantity":        trade.Quantity,
		"price":           trade.Price,
		"value":           trade.Value,
		"commission":      trade.Commission,
		"pnl":             trade.PnL,
		"strategy":        trade.Strategy,
		"order_id":        trade.OrderID,
		"timestamp":       trade.Timestamp,
		"execution_time":  trade.ExecutionTime,
	}

	req := &models.IndexRequest{
		Index:    "trades",
		ID:       trade.ID,
		Document: doc,
	}

	if err := i.Index(req); err != nil {
		i.logger.Error("Failed to queue trade for indexing",
			zap.Error(err),
			zap.String("trade_id", trade.ID),
		)
	}
}

func (i *Indexer) handleOrderMessage(msg *nats.Msg) {
	var order models.Order
	if err := json.Unmarshal(msg.Data, &order); err != nil {
		i.logger.Error("Failed to unmarshal order message", zap.Error(err))
		return
	}

	doc := map[string]interface{}{
		"id":               order.ID,
		"symbol":           order.Symbol,
		"side":             order.Side,
		"type":             order.Type,
		"status":           order.Status,
		"quantity":         order.Quantity,
		"price":            order.Price,
		"filled_quantity":  order.FilledQuantity,
		"avg_fill_price":   order.AvgFillPrice,
		"strategy":         order.Strategy,
		"time_in_force":    order.TimeInForce,
		"created_at":       order.CreatedAt,
		"updated_at":       order.UpdatedAt,
	}

	if order.CancelledAt != nil {
		doc["cancelled_at"] = order.CancelledAt
	}

	req := &models.IndexRequest{
		Index:    "orders",
		ID:       order.ID,
		Document: doc,
	}

	if err := i.Index(req); err != nil {
		i.logger.Error("Failed to queue order for indexing",
			zap.Error(err),
			zap.String("order_id", order.ID),
		)
	}
}

func (i *Indexer) handleStrategyMessage(msg *nats.Msg) {
	var strategy models.Strategy
	if err := json.Unmarshal(msg.Data, &strategy); err != nil {
		i.logger.Error("Failed to unmarshal strategy message", zap.Error(err))
		return
	}

	doc := map[string]interface{}{
		"id":          strategy.ID,
		"name":        strategy.Name,
		"type":        strategy.Type,
		"status":      strategy.Status,
		"symbols":     strategy.Symbols,
		"parameters":  strategy.Parameters,
		"performance": strategy.Performance,
		"created_at":  strategy.CreatedAt,
		"updated_at":  strategy.UpdatedAt,
	}

	req := &models.IndexRequest{
		Index:    "strategies",
		ID:       strategy.ID,
		Document: doc,
	}

	if err := i.Index(req); err != nil {
		i.logger.Error("Failed to queue strategy for indexing",
			zap.Error(err),
			zap.String("strategy_id", strategy.ID),
		)
	}
}

func (i *Indexer) handleMarketDataMessage(msg *nats.Msg) {
	var marketData models.MarketData
	if err := json.Unmarshal(msg.Data, &marketData); err != nil {
		i.logger.Error("Failed to unmarshal market data message", zap.Error(err))
		return
	}

	doc := map[string]interface{}{
		"symbol":    marketData.Symbol,
		"timestamp": marketData.Timestamp,
		"open":      marketData.Open,
		"high":      marketData.High,
		"low":       marketData.Low,
		"close":     marketData.Close,
		"volume":    marketData.Volume,
		"vwap":      marketData.VWAP,
		"trades":    marketData.Trades,
	}

	// Generate ID from symbol and timestamp
	id := fmt.Sprintf("%s-%d", marketData.Symbol, marketData.Timestamp.Unix())

	req := &models.IndexRequest{
		Index:    "market_data",
		ID:       id,
		Document: doc,
	}

	if err := i.Index(req); err != nil {
		i.logger.Error("Failed to queue market data for indexing",
			zap.Error(err),
			zap.String("symbol", marketData.Symbol),
		)
	}
}

func (i *Indexer) handleLogMessage(msg *nats.Msg) {
	var logEntry models.LogEntry
	if err := json.Unmarshal(msg.Data, &logEntry); err != nil {
		i.logger.Error("Failed to unmarshal log message", zap.Error(err))
		return
	}

	doc := map[string]interface{}{
		"id":        logEntry.ID,
		"level":     logEntry.Level,
		"service":   logEntry.Service,
		"message":   logEntry.Message,
		"timestamp": logEntry.Timestamp,
	}

	if len(logEntry.Fields) > 0 {
		doc["fields"] = logEntry.Fields
	}
	if logEntry.TraceID != "" {
		doc["trace_id"] = logEntry.TraceID
	}
	if logEntry.SpanID != "" {
		doc["span_id"] = logEntry.SpanID
	}

	req := &models.IndexRequest{
		Index:    "logs",
		ID:       logEntry.ID,
		Document: doc,
	}

	if err := i.Index(req); err != nil {
		i.logger.Error("Failed to queue log for indexing",
			zap.Error(err),
			zap.String("log_id", logEntry.ID),
		)
	}
}
