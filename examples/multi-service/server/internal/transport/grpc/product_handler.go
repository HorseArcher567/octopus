package grpc

import (
	"context"

	"github.com/HorseArcher567/octopus/examples/multi-service/proto/pb"
	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/service"
	"github.com/HorseArcher567/octopus/pkg/xlog"
)

type ProductHandler struct {
	pb.UnimplementedProductServer
	svc *service.ProductService
	log *xlog.Logger
}

func NewProductHandler(svc *service.ProductService, log *xlog.Logger) *ProductHandler {
	return &ProductHandler{svc: svc, log: log}
}

func (h *ProductHandler) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	log := xlog.GetOr(ctx, h.log).With("product_id", req.ProductId)
	log.Info("get product")

	product, err := h.svc.GetByID(ctx, req.ProductId)
	if err != nil {
		log.Error("get product failed", "error", err)
		return nil, mapError(err, "product not found")
	}

	return &pb.GetProductResponse{
		ProductId:   product.ProductID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       int32(product.Stock),
	}, nil
}

func (h *ProductHandler) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	log := xlog.GetOr(ctx, h.log).With("page", req.Page, "page_size", req.PageSize)
	log.Info("list products")

	products, total, err := h.svc.List(ctx, req.Page, req.PageSize)
	if err != nil {
		log.Error("list products failed", "error", err)
		return nil, mapError(err, "failed to list products")
	}

	resp := make([]*pb.GetProductResponse, 0, len(products))
	for _, p := range products {
		resp = append(resp, &pb.GetProductResponse{
			ProductId:   p.ProductID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Stock:       int32(p.Stock),
		})
	}

	return &pb.ListProductsResponse{Products: resp, Total: int32(total)}, nil
}
