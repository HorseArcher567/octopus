package grpc

import (
	"context"

	"github.com/HorseArcher567/octopus/examples/multi-service/proto/pb"
	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/service"
	"github.com/HorseArcher567/octopus/pkg/xlog"
)

type OrderHandler struct {
	pb.UnimplementedOrderServer
	svc *service.OrderService
	log *xlog.Logger
}

func NewOrderHandler(svc *service.OrderService, log *xlog.Logger) *OrderHandler {
	return &OrderHandler{svc: svc, log: log}
}

func (h *OrderHandler) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	log := xlog.GetOr(ctx, h.log).With("order_id", req.OrderId)
	log.Info("get order")

	order, err := h.svc.GetByID(ctx, req.OrderId)
	if err != nil {
		log.Error("get order failed", "error", err)
		return nil, mapError(err, "order not found")
	}

	return &pb.GetOrderResponse{
		OrderId:     order.OrderID,
		UserId:      order.UserID,
		ProductName: order.ProductName,
		Amount:      order.Amount,
		Status:      order.Status,
	}, nil
}

func (h *OrderHandler) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	log := xlog.GetOr(ctx, h.log).With("user_id", req.UserId, "product", req.ProductName, "amount", req.Amount)
	log.Info("create order")

	id, err := h.svc.Create(ctx, req.UserId, req.ProductName, req.Amount)
	if err != nil {
		log.Error("create order failed", "error", err)
		return nil, mapError(err, "failed to create order")
	}

	return &pb.CreateOrderResponse{OrderId: id, Message: "Order created successfully"}, nil
}
