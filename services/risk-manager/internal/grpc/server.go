package grpc

import (
	"context"
	"time"

	"github.com/b25/services/risk-manager/internal/cache"
	"github.com/b25/services/risk-manager/internal/emergency"
	"github.com/b25/services/risk-manager/internal/limits"
	"github.com/b25/services/risk-manager/internal/risk"
	pb "github.com/b25/services/risk-manager/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RiskServer implements the gRPC RiskManager service
type RiskServer struct {
	pb.UnimplementedRiskManagerServer
	logger         *zap.Logger
	calculator     *risk.Calculator
	policyEngine   *limits.PolicyEngine
	policyCache    *cache.PolicyCache
	priceCache     *cache.MarketPriceCache
	stopManager    *emergency.StopManager
}

// NewRiskServer creates a new gRPC risk server
func NewRiskServer(
	logger *zap.Logger,
	calculator *risk.Calculator,
	policyEngine *limits.PolicyEngine,
	policyCache *cache.PolicyCache,
	priceCache *cache.MarketPriceCache,
	stopManager *emergency.StopManager,
) *RiskServer {
	return &RiskServer{
		logger:       logger,
		calculator:   calculator,
		policyEngine: policyEngine,
		policyCache:  policyCache,
		priceCache:   priceCache,
		stopManager:  stopManager,
	}
}

// CheckOrder performs pre-trade risk validation for an order
func (s *RiskServer) CheckOrder(ctx context.Context, req *pb.OrderRiskRequest) (*pb.OrderRiskResponse, error) {
	startTime := time.Now()

	// Check if emergency stop is active
	if s.stopManager.ShouldBlockOrders() {
		return &pb.OrderRiskResponse{
			Approved:          false,
			Violations:        []string{"Emergency stop is active - all trading halted"},
			RejectionReason:   "emergency_stop_active",
			ProcessingTimeUs:  time.Since(startTime).Microseconds(),
		}, nil
	}

	// Get current price
	currentPrice := req.Price
	if currentPrice == 0 {
		// Market order - need current price
		price, err := s.priceCache.GetPrice(ctx, req.Symbol)
		if err != nil {
			s.logger.Error("failed to get current price",
				zap.String("symbol", req.Symbol),
				zap.Error(err),
			)
			return nil, status.Errorf(codes.Unavailable, "price data unavailable for %s", req.Symbol)
		}
		currentPrice = price
	}

	// Create mock account state (in production, fetch from Account Monitor)
	accountState := s.getMockAccountState(req.AccountId)

	// Create order for simulation
	order := risk.Order{
		OrderID:    req.OrderId,
		Symbol:     req.Symbol,
		Side:       req.Side,
		Quantity:   req.Quantity,
		Price:      req.Price,
		OrderType:  req.OrderType,
		StrategyID: req.StrategyId,
	}

	// Simulate order impact on account
	postTradeState, err := s.calculator.SimulateOrder(accountState, order, currentPrice)
	if err != nil {
		return &pb.OrderRiskResponse{
			Approved:          false,
			Violations:        []string{err.Error()},
			RejectionReason:   "simulation_failed",
			ProcessingTimeUs:  time.Since(startTime).Microseconds(),
		}, nil
	}

	// Calculate post-trade risk metrics
	postTradeMetrics := s.calculator.CalculateMetrics(postTradeState)

	// Convert metrics to map for policy evaluation
	metricsMap := limits.MetricsFromRiskMetrics(
		postTradeMetrics.Leverage,
		postTradeMetrics.MarginRatio,
		postTradeMetrics.DrawdownDaily,
		postTradeMetrics.DrawdownMax,
		postTradeMetrics.PositionConcentration,
	)

	// Evaluate policies
	violations := s.policyEngine.EvaluateAll(metricsMap, req.Symbol, req.StrategyId)

	// Filter only hard and emergency violations (soft violations are warnings only)
	blockingViolations := make([]*limits.PolicyViolation, 0)
	for _, v := range violations {
		if v.Policy.Type == limits.PolicyTypeHard || v.Policy.Type == limits.PolicyTypeEmergency {
			blockingViolations = append(blockingViolations, v)
		}
	}

	approved := len(blockingViolations) == 0
	processingTime := time.Since(startTime).Microseconds()

	// Log the check
	s.logger.Info("order risk check",
		zap.String("order_id", req.OrderId),
		zap.String("symbol", req.Symbol),
		zap.Bool("approved", approved),
		zap.Int("violations", len(blockingViolations)),
		zap.Int64("processing_time_us", processingTime),
	)

	// Build response
	response := &pb.OrderRiskResponse{
		Approved:          approved,
		Violations:        limits.FormatViolations(blockingViolations),
		PostTradeMetrics:  s.metricsToProto(postTradeMetrics),
		ProcessingTimeUs:  processingTime,
	}

	if !approved {
		response.RejectionReason = "policy_violation"
	}

	return response, nil
}

// CheckOrderBatch performs batch validation
func (s *RiskServer) CheckOrderBatch(ctx context.Context, req *pb.BatchOrderRiskRequest) (*pb.BatchOrderRiskResponse, error) {
	startTime := time.Now()

	responses := make([]*pb.OrderRiskResponse, len(req.Orders))
	for i, order := range req.Orders {
		resp, err := s.CheckOrder(ctx, order)
		if err != nil {
			return nil, err
		}
		responses[i] = resp
	}

	return &pb.BatchOrderRiskResponse{
		Responses:             responses,
		TotalProcessingTimeUs: time.Since(startTime).Microseconds(),
	}, nil
}

// GetRiskMetrics returns current risk metrics
func (s *RiskServer) GetRiskMetrics(ctx context.Context, req *pb.RiskMetricsRequest) (*pb.RiskMetricsResponse, error) {
	// Get account state
	accountState := s.getMockAccountState(req.AccountId)

	// Calculate metrics
	metrics := s.calculator.CalculateMetrics(accountState)

	return &pb.RiskMetricsResponse{
		Metrics:   s.metricsToProto(metrics),
		Timestamp: time.Now().Unix(),
	}, nil
}

// TriggerEmergencyStop triggers an emergency stop
func (s *RiskServer) TriggerEmergencyStop(ctx context.Context, req *pb.EmergencyStopRequest) (*pb.EmergencyStopResponse, error) {
	s.logger.Warn("emergency stop triggered via RPC",
		zap.String("reason", req.Reason),
		zap.String("triggered_by", req.TriggeredBy),
	)

	err := s.stopManager.Trigger(ctx, req.Reason, req.TriggeredBy, req.Force)
	if err != nil && !req.Force {
		return &pb.EmergencyStopResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	status := s.stopManager.GetStatus()

	return &pb.EmergencyStopResponse{
		Success: true,
		Message: "Emergency stop activated",
		Status:  s.stopStatusToProto(status),
	}, nil
}

// GetEmergencyStopStatus returns current emergency stop status
func (s *RiskServer) GetEmergencyStopStatus(ctx context.Context, req *pb.EmergencyStopStatusRequest) (*pb.EmergencyStopStatusResponse, error) {
	status := s.stopManager.GetStatus()
	return &pb.EmergencyStopStatusResponse{
		Status: s.stopStatusToProto(status),
	}, nil
}

// ReEnableTrading re-enables trading after emergency stop
func (s *RiskServer) ReEnableTrading(ctx context.Context, req *pb.ReEnableTradingRequest) (*pb.ReEnableTradingResponse, error) {
	s.logger.Info("re-enabling trading after emergency stop",
		zap.String("authorized_by", req.AuthorizedBy),
		zap.String("reason", req.Reason),
	)

	err := s.stopManager.ReEnable(req.AuthorizedBy, req.Reason)
	if err != nil {
		return &pb.ReEnableTradingResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.ReEnableTradingResponse{
		Success:     true,
		Message:     "Trading re-enabled successfully",
		ReEnabledAt: time.Now().Unix(),
	}, nil
}

// metricsToProto converts RiskMetrics to protobuf format
func (s *RiskServer) metricsToProto(metrics risk.RiskMetrics) *pb.RiskMetrics {
	// Calculate limit utilization (mock data)
	limitUtilization := make(map[string]float64)
	limitUtilization["leverage"] = metrics.Leverage / 10.0 // Assuming max leverage is 10x
	limitUtilization["drawdown"] = metrics.DrawdownMax / 0.20 // Assuming max drawdown is 20%

	return &pb.RiskMetrics{
		MarginRatio:           metrics.MarginRatio,
		Leverage:              metrics.Leverage,
		DrawdownDaily:         metrics.DrawdownDaily,
		DrawdownMax:           metrics.DrawdownMax,
		DailyPnl:              metrics.DailyPnL,
		UnrealizedPnl:         metrics.UnrealizedPnL,
		TotalEquity:           metrics.TotalEquity,
		TotalMarginUsed:       metrics.TotalMarginUsed,
		PositionConcentration: metrics.PositionConcentration,
		LimitUtilization:      limitUtilization,
		OpenPositions:         int32(metrics.OpenPositions),
		PendingOrders:         int32(metrics.PendingOrders),
	}
}

// stopStatusToProto converts StopStatus to protobuf format
func (s *RiskServer) stopStatusToProto(status emergency.StopStatus) *pb.EmergencyStopStatus {
	return &pb.EmergencyStopStatus{
		IsStopped:       status.IsStopped,
		StoppedAt:       status.StoppedAt.Unix(),
		StopReason:      status.StopReason,
		TriggeredBy:     status.TriggeredBy,
		OrdersCanceled:  int32(status.OrdersCanceled),
		PositionsClosed: int32(status.PositionsClosed),
		Completed:       status.Completed,
		CompletedAt:     status.CompletedAt.Unix(),
	}
}

// getMockAccountState returns mock account state (replace with real Account Monitor client)
func (s *RiskServer) getMockAccountState(accountID string) risk.AccountState {
	return risk.AccountState{
		Equity:           100000.0,
		Balance:          100000.0,
		UnrealizedPnL:    0.0,
		MarginUsed:       10000.0,
		AvailableMargin:  90000.0,
		Positions:        []risk.Position{},
		PendingOrders:    []risk.Order{},
		PeakEquity:       105000.0,
		DailyStartEquity: 98000.0,
	}
}
