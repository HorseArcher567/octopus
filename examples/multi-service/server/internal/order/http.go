package order

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	"github.com/gin-gonic/gin"
)

type HTTPHandler struct {
	svc *Service
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
type errorResponse struct {
	Error string `json:"error"`
}

func NewHTTPHandler(svc *Service, log *xlog.Logger) *HTTPHandler {
	return &HTTPHandler{svc: svc, log: log}
}

func RegisterHTTP(engine *api.Engine, h *HTTPHandler) {
	engine.GET("/orders/:id", h.GetOrder)
	engine.POST("/orders", h.CreateOrder)
}

func (h *HTTPHandler) GetOrder(c *gin.Context) {
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
		writeHTTPError(c, err, "order not found")
		return
	}
	c.JSON(http.StatusOK, getOrderResponse{OrderID: order.OrderID, UserID: order.UserID, ProductName: order.ProductName, Amount: order.Amount, Status: order.Status})
}

func (h *HTTPHandler) CreateOrder(c *gin.Context) {
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
		writeHTTPError(c, err, "failed to create order")
		return
	}
	c.JSON(http.StatusOK, createOrderResponse{OrderID: id, Message: "Order created successfully"})
}

func writeHTTPError(c *gin.Context, err error, notFoundMsg string) {
	if err == nil {
		return
	}
	switch {
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, errorResponse{Error: notFoundMsg})
	case errors.Is(err, ErrInvalidArgument):
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
	}
}
