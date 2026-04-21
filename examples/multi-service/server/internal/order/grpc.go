package order

import (
	"context"
	"errors"

	"github.com/HorseArcher567/octopus/examples/multi-service/proto/pb"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCHandler struct { pb.UnimplementedOrderServer; svc *Service; log *xlog.Logger }

func NewGRPCHandler(svc *Service, log *xlog.Logger) *GRPCHandler { return &GRPCHandler{svc: svc, log: log} }
func RegisterGRPC(s *grpc.Server, h *GRPCHandler) { pb.RegisterOrderServer(s, h) }

func (h *GRPCHandler) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	log := xlog.GetOr(ctx, h.log).With("order_id", req.OrderId); log.Info("get order")
	order, err := h.svc.GetByID(ctx, req.OrderId)
	if err != nil { log.Error("get order failed", "error", err); return nil, mapGRPCError(err, "order not found") }
	return &pb.GetOrderResponse{OrderId: order.OrderID, UserId: order.UserID, ProductName: order.ProductName, Amount: order.Amount, Status: order.Status}, nil
}

func (h *GRPCHandler) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	log := xlog.GetOr(ctx, h.log).With("user_id", req.UserId, "product", req.ProductName, "amount", req.Amount); log.Info("create order")
	id, err := h.svc.Create(ctx, req.UserId, req.ProductName, req.Amount)
	if err != nil { log.Error("create order failed", "error", err); return nil, mapGRPCError(err, "failed to create order") }
	return &pb.CreateOrderResponse{OrderId: id, Message: "Order created successfully"}, nil
}

func mapGRPCError(err error, notFoundMsg string) error {
	if err == nil { return nil }
	if errors.Is(err, ErrNotFound) { return status.Error(codes.NotFound, notFoundMsg) }
	if errors.Is(err, ErrInvalidArgument) { return status.Error(codes.InvalidArgument, err.Error()) }
	return status.Error(codes.Internal, err.Error())
}
