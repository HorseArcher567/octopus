package http

import (
	"net/http"
	"strconv"

	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/service"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	svc *service.ProductService
	log *xlog.Logger
}

type getProductResponse struct {
	ProductID   int64   `json:"product_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int32   `json:"stock"`
}

type listProductsResponse struct {
	Products []getProductResponse `json:"products"`
	Total    int32                `json:"total"`
}

func NewProductHandler(svc *service.ProductService, log *xlog.Logger) *ProductHandler {
	return &ProductHandler{svc: svc, log: log}
}

func (h *ProductHandler) GetProduct(c *gin.Context) {
	productID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid product id"})
		return
	}

	ctx := c.Request.Context()
	log := xlog.GetOr(ctx, h.log).With("product_id", productID)
	log.Info("get product")

	product, err := h.svc.GetByID(ctx, productID)
	if err != nil {
		log.Error("get product failed", "error", err)
		writeError(c, err, "product not found")
		return
	}

	c.JSON(http.StatusOK, getProductResponse{
		ProductID:   product.ProductID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       int32(product.Stock),
	})
}

func (h *ProductHandler) ListProducts(c *gin.Context) {
	page, err := parseQueryInt32(c.Query("page"), 1)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid page"})
		return
	}
	pageSize, err := parseQueryInt32(c.Query("page_size"), 10)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid page_size"})
		return
	}

	ctx := c.Request.Context()
	log := xlog.GetOr(ctx, h.log).With("page", page, "page_size", pageSize)
	log.Info("list products")

	products, total, err := h.svc.List(ctx, page, pageSize)
	if err != nil {
		log.Error("list products failed", "error", err)
		writeError(c, err, "failed to list products")
		return
	}

	resp := make([]getProductResponse, 0, len(products))
	for _, product := range products {
		resp = append(resp, getProductResponse{
			ProductID:   product.ProductID,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			Stock:       int32(product.Stock),
		})
	}

	c.JSON(http.StatusOK, listProductsResponse{
		Products: resp,
		Total:    int32(total),
	})
}

func parseQueryInt32(value string, defaultValue int32) (int32, error) {
	if value == "" {
		return defaultValue, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(parsed), nil
}
