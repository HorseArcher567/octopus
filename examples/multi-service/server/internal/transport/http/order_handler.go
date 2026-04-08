package http

import (
	"net/http"
	"strconv"

	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/service"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	svc *service.OrderService
	log *xlog.Logger
}

type createOrderRequest struct {
	UserID      int64   `json:"user_id"`
	ProductName string  `json:"product_name"`
	Amount      float64 `json:"amount"`
}

type getOrderResponse struct {
	OrderID     int64   `json:"order_id"`
	UserID      int64   `json:"user_id"`
	ProductName string  `json:"product_name"`
	Amount      float64 `json:"amount"`
	Status      string  `json:"status"`
}

type createOrderResponse struct {
	OrderID int64  `json:"order_id"`
	Message string `json:"message"`
}

func NewOrderHandler(svc *service.OrderService, log *xlog.Logger) *OrderHandler {
	return &OrderHandler{svc: svc, log: log}
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid order id"})
		return
	}

	ctx := c.Request.Context()
	log := xlog.GetOr(ctx, h.log).With("order_id", orderID)
	log.Info("get order")

	order, err := h.svc.GetByID(ctx, orderID)
	if err != nil {
		log.Error("get order failed", "error", err)
		writeError(c, err, "order not found")
		return
	}

	c.JSON(http.StatusOK, getOrderResponse{
		OrderID:     order.OrderID,
		UserID:      order.UserID,
		ProductName: order.ProductName,
		Amount:      order.Amount,
		Status:      order.Status,
	})
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	ctx := c.Request.Context()
	log := xlog.GetOr(ctx, h.log).With("user_id", req.UserID, "product", req.ProductName, "amount", req.Amount)
	log.Info("create order")

	id, err := h.svc.Create(ctx, req.UserID, req.ProductName, req.Amount)
	if err != nil {
		log.Error("create order failed", "error", err)
		writeError(c, err, "failed to create order")
		return
	}

	c.JSON(http.StatusOK, createOrderResponse{
		OrderID: id,
		Message: "Order created successfully",
	})
}
