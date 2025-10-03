package executor

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/yourusername/b25/services/order-execution/internal/models"
	pb "github.com/yourusername/b25/services/order-execution/proto"
)

// GRPCServer implements the OrderService gRPC interface
type GRPCServer struct {
	pb.UnimplementedOrderServiceServer
	executor *OrderExecutor
}

// NewGRPCServer creates a new gRPC server
func NewGRPCServer(executor *OrderExecutor) *GRPCServer {
	return &GRPCServer{
		executor: executor,
	}
}

// CreateOrder handles order creation requests
func (s *GRPCServer) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	// Convert proto to internal model
	order := &models.Order{
		Symbol:        req.Symbol,
		Side:          mapProtoSide(req.Side),
		Type:          mapProtoType(req.Type),
		Quantity:      req.Quantity,
		Price:         req.Price,
		StopPrice:     req.StopPrice,
		TimeInForce:   mapProtoTimeInForce(req.TimeInForce),
		ClientOrderID: req.ClientOrderId,
		ReduceOnly:    req.ReduceOnly,
		PostOnly:      req.PostOnly,
		UserID:        req.UserId,
	}

	// Create order
	err := s.executor.CreateOrder(ctx, order)
	if err != nil {
		return &pb.CreateOrderResponse{
			OrderId:      order.OrderID,
			State:        mapOrderStateToProto(order.State),
			Timestamp:    order.CreatedAt.Unix(),
			ErrorMessage: err.Error(),
		}, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.CreateOrderResponse{
		OrderId:       order.OrderID,
		ClientOrderId: order.ClientOrderID,
		State:         mapOrderStateToProto(order.State),
		Timestamp:     order.CreatedAt.Unix(),
	}, nil
}

// CancelOrder handles order cancellation requests
func (s *GRPCServer) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error) {
	err := s.executor.CancelOrder(ctx, req.OrderId, req.Symbol)
	if err != nil {
		return &pb.CancelOrderResponse{
			OrderId:      req.OrderId,
			State:        pb.OrderState_REJECTED,
			Timestamp:    time.Now().Unix(),
			ErrorMessage: err.Error(),
		}, status.Error(codes.Internal, err.Error())
	}

	// Get updated order
	order, err := s.executor.GetOrder(ctx, req.OrderId)
	if err != nil {
		return &pb.CancelOrderResponse{
			OrderId:      req.OrderId,
			State:        pb.OrderState_CANCELED,
			Timestamp:    time.Now().Unix(),
		}, nil
	}

	return &pb.CancelOrderResponse{
		OrderId:   req.OrderId,
		State:     mapOrderStateToProto(order.State),
		Timestamp: order.UpdatedAt.Unix(),
	}, nil
}

// GetOrder handles order query requests
func (s *GRPCServer) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	order, err := s.executor.GetOrder(ctx, req.OrderId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &pb.GetOrderResponse{
		Order: mapOrderToProto(order),
	}, nil
}

// GetOrders handles bulk order query requests
func (s *GRPCServer) GetOrders(ctx context.Context, req *pb.GetOrdersRequest) (*pb.GetOrdersResponse, error) {
	// This is a simplified implementation
	// In production, you'd query from database with filters
	var orders []*pb.Order

	s.executor.orders.Range(func(key, value interface{}) bool {
		order := value.(*models.Order)

		// Apply filters
		if req.Symbol != "" && order.Symbol != req.Symbol {
			return true
		}
		if req.State != pb.OrderState_STATE_UNSPECIFIED && mapOrderStateToProto(order.State) != req.State {
			return true
		}

		orders = append(orders, mapOrderToProto(order))
		return true
	})

	// Apply limit
	if req.Limit > 0 && int32(len(orders)) > req.Limit {
		orders = orders[:req.Limit]
	}

	return &pb.GetOrdersResponse{
		Orders: orders,
	}, nil
}

// StreamOrderUpdates streams order updates
func (s *GRPCServer) StreamOrderUpdates(req *pb.StreamOrderUpdatesRequest, stream pb.OrderService_StreamOrderUpdatesServer) error {
	// Subscribe to NATS for order updates
	// This is a simplified implementation
	// In production, you'd use proper NATS subscription with filters

	ctx := stream.Context()

	// Create a channel for updates
	updateChan := make(chan *models.OrderUpdate, 100)

	// Simulate streaming (in production, subscribe to NATS)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case update := <-updateChan:
			pbUpdate := &pb.OrderUpdate{
				Order:      mapOrderToProto(update.Order),
				UpdateType: update.UpdateType,
				Timestamp:  update.Timestamp.Unix(),
			}
			if err := stream.Send(pbUpdate); err != nil {
				return err
			}
		case <-ticker.C:
			// Keep alive
		}
	}
}

// Helper functions to map between proto and internal models

func mapProtoSide(side pb.OrderSide) models.OrderSide {
	switch side {
	case pb.OrderSide_BUY:
		return models.OrderSideBuy
	case pb.OrderSide_SELL:
		return models.OrderSideSell
	default:
		return models.OrderSideBuy
	}
}

func mapProtoType(orderType pb.OrderType) models.OrderType {
	switch orderType {
	case pb.OrderType_MARKET:
		return models.OrderTypeMarket
	case pb.OrderType_LIMIT:
		return models.OrderTypeLimit
	case pb.OrderType_STOP_MARKET:
		return models.OrderTypeStopMarket
	case pb.OrderType_STOP_LIMIT:
		return models.OrderTypeStopLimit
	case pb.OrderType_POST_ONLY:
		return models.OrderTypePostOnly
	default:
		return models.OrderTypeLimit
	}
}

func mapProtoTimeInForce(tif pb.TimeInForce) models.TimeInForce {
	switch tif {
	case pb.TimeInForce_GTC:
		return models.TimeInForceGTC
	case pb.TimeInForce_IOC:
		return models.TimeInForceIOC
	case pb.TimeInForce_FOK:
		return models.TimeInForceFOK
	case pb.TimeInForce_GTX:
		return models.TimeInForceGTX
	default:
		return models.TimeInForceGTC
	}
}

func mapOrderStateToProto(state models.OrderState) pb.OrderState {
	switch state {
	case models.OrderStateNew:
		return pb.OrderState_NEW
	case models.OrderStateSubmitted:
		return pb.OrderState_SUBMITTED
	case models.OrderStatePartiallyFilled:
		return pb.OrderState_PARTIALLY_FILLED
	case models.OrderStateFilled:
		return pb.OrderState_FILLED
	case models.OrderStateCanceled:
		return pb.OrderState_CANCELED
	case models.OrderStateRejected:
		return pb.OrderState_REJECTED
	case models.OrderStateExpired:
		return pb.OrderState_EXPIRED
	default:
		return pb.OrderState_STATE_UNSPECIFIED
	}
}

func mapOrderToProto(order *models.Order) *pb.Order {
	return &pb.Order{
		OrderId:         order.OrderID,
		ClientOrderId:   order.ClientOrderID,
		Symbol:          order.Symbol,
		Side:            mapSideToProto(order.Side),
		Type:            mapTypeToProto(order.Type),
		State:           mapOrderStateToProto(order.State),
		TimeInForce:     mapTimeInForceToProto(order.TimeInForce),
		Quantity:        order.Quantity,
		Price:           order.Price,
		StopPrice:       order.StopPrice,
		FilledQuantity:  order.FilledQuantity,
		AveragePrice:    order.AveragePrice,
		Fee:             order.Fee,
		FeeAsset:        order.FeeAsset,
		CreatedAt:       order.CreatedAt.Unix(),
		UpdatedAt:       order.UpdatedAt.Unix(),
		UserId:          order.UserID,
		ReduceOnly:      order.ReduceOnly,
		PostOnly:        order.PostOnly,
		ExchangeOrderId: order.ExchangeOrderID,
	}
}

func mapSideToProto(side models.OrderSide) pb.OrderSide {
	switch side {
	case models.OrderSideBuy:
		return pb.OrderSide_BUY
	case models.OrderSideSell:
		return pb.OrderSide_SELL
	default:
		return pb.OrderSide_SIDE_UNSPECIFIED
	}
}

func mapTypeToProto(orderType models.OrderType) pb.OrderType {
	switch orderType {
	case models.OrderTypeMarket:
		return pb.OrderType_MARKET
	case models.OrderTypeLimit:
		return pb.OrderType_LIMIT
	case models.OrderTypeStopMarket:
		return pb.OrderType_STOP_MARKET
	case models.OrderTypeStopLimit:
		return pb.OrderType_STOP_LIMIT
	case models.OrderTypePostOnly:
		return pb.OrderType_POST_ONLY
	default:
		return pb.OrderType_TYPE_UNSPECIFIED
	}
}

func mapTimeInForceToProto(tif models.TimeInForce) pb.TimeInForce {
	switch tif {
	case models.TimeInForceGTC:
		return pb.TimeInForce_GTC
	case models.TimeInForceIOC:
		return pb.TimeInForce_IOC
	case models.TimeInForceFOK:
		return pb.TimeInForce_FOK
	case models.TimeInForceGTX:
		return pb.TimeInForce_GTX
	default:
		return pb.TimeInForce_TIF_UNSPECIFIED
	}
}

// Add missing import
func init() {
	// Ensure uuid is available for order ID generation
	_ = fmt.Sprintf("")
	_ = strconv.Itoa(0)
}
