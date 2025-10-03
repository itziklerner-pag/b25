package grpcserver

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/yourorg/b25/services/account-monitor/internal/metrics"
	"github.com/yourorg/b25/services/account-monitor/internal/monitor"
	pb "github.com/yourorg/b25/services/account-monitor/pkg/proto"
)

// RegisterAccountMonitorServer registers the gRPC server
// This is a placeholder - actual proto definitions would be needed
func RegisterAccountMonitorServer(s *grpc.Server, monitor *monitor.AccountMonitor) {
	// pb.RegisterAccountMonitorServer(s, &server{monitor: monitor})
	// For now, we'll create a simple implementation
}

type server struct {
	pb.UnimplementedAccountMonitorServer
	monitor *monitor.AccountMonitor
}

// GetAccountState returns current account state
func (s *server) GetAccountState(ctx context.Context, req *pb.AccountRequest) (*pb.AccountState, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		metrics.GRPCDuration.WithLabelValues("GetAccountState").Observe(duration.Seconds())
	}()

	// Implementation would go here
	metrics.GRPCRequests.WithLabelValues("GetAccountState", "success").Inc()

	return &pb.AccountState{
		Timestamp: timestamppb.Now(),
	}, nil
}

// GetPosition returns position for a symbol
func (s *server) GetPosition(ctx context.Context, req *pb.PositionRequest) (*pb.Position, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		metrics.GRPCDuration.WithLabelValues("GetPosition").Observe(duration.Seconds())
	}()

	// TODO: Implement actual position retrieval
	// pos, err := s.monitor.GetPositionData(req.Symbol)
	// if err != nil {
	// 	metrics.GRPCRequests.WithLabelValues("GetPosition", "error").Inc()
	// 	return nil, status.Errorf(codes.NotFound, "position not found: %v", err)
	// }

	metrics.GRPCRequests.WithLabelValues("GetPosition", "success").Inc()

	return &pb.Position{
		Symbol:    req.Symbol,
		Quantity:  "0",
		Timestamp: timestamppb.Now(),
	}, nil
}

// Helper to convert internal position to proto (placeholder)
func convertPosition(pos interface{}) *pb.Position {
	return &pb.Position{
		Symbol:    "BTCUSDT",
		Quantity:  "0",
		Timestamp: timestamppb.Now(),
	}
}

func decimalToString(d decimal.Decimal) string {
	return d.String()
}
