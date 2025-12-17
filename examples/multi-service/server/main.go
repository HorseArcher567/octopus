package main

import (
	"context"

	"github.com/HorseArcher567/octopus/examples/multi-service/proto/pb"
	"github.com/HorseArcher567/octopus/pkg/logger"
	"github.com/HorseArcher567/octopus/pkg/rpc"
	"google.golang.org/grpc"
)

// UserServiceImpl 用户服务实现
type UserServiceImpl struct {
	pb.UnimplementedUserServiceServer
}

func (s *UserServiceImpl) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	logger.Info("get user called", "user_id", req.UserId)
	return &pb.GetUserResponse{
		UserId:   req.UserId,
		Username: "testuser",
		Email:    "test@example.com",
	}, nil
}

func (s *UserServiceImpl) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	logger.Info("create user called",
		"username", req.Username,
		"email", req.Email,
	)
	return &pb.CreateUserResponse{
		UserId:  1001,
		Message: "User created successfully",
	}, nil
}

// OrderServiceImpl 订单服务实现
type OrderServiceImpl struct {
	pb.UnimplementedOrderServiceServer
}

func (s *OrderServiceImpl) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	logger.Info("get order called", "order_id", req.OrderId)
	return &pb.GetOrderResponse{
		OrderId:     req.OrderId,
		UserId:      1001,
		ProductName: "Sample Product",
		Amount:      99.99,
		Status:      "completed",
	}, nil
}

func (s *OrderServiceImpl) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	logger.Info("create order called",
		"user_id", req.UserId,
		"product", req.ProductName,
		"amount", req.Amount,
	)
	return &pb.CreateOrderResponse{
		OrderId: 2001,
		Message: "Order created successfully",
	}, nil
}

// ProductServiceImpl 产品服务实现
type ProductServiceImpl struct {
	pb.UnimplementedProductServiceServer
}

func (s *ProductServiceImpl) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	logger.Info("get product called", "product_id", req.ProductId)
	return &pb.GetProductResponse{
		ProductId:   req.ProductId,
		Name:        "Sample Product",
		Description: "This is a sample product",
		Price:       99.99,
		Stock:       100,
	}, nil
}

func (s *ProductServiceImpl) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	logger.Info("list products called",
		"page", req.Page,
		"page_size", req.PageSize,
	)
	return &pb.ListProductsResponse{
		Products: []*pb.GetProductResponse{
			{
				ProductId:   1,
				Name:        "Product 1",
				Description: "First product",
				Price:       49.99,
				Stock:       50,
			},
			{
				ProductId:   2,
				Name:        "Product 2",
				Description: "Second product",
				Price:       79.99,
				Stock:       30,
			},
		},
		Total: 2,
	}, nil
}

func main() {
	// 初始化日志
	logger.Init(&logger.Config{
		Level:     "debug",
		Format:    "text",
		AddSource: true,
	})

	// 配置服务器
	config := &rpc.ServerConfig{
		AppName:          "multi-service-demo",
		Host:             "127.0.0.1",
		Port:             9000,
		EtcdAddr:         []string{"localhost:2379"},
		TTL:              10,
		EnableReflection: true, // 开启反射，便于使用 grpcurl/grpcui 调试
	}

	// 创建 RPC 服务器
	server := rpc.NewServer(config)

	// 注册多个服务
	server.RegisterService(func(s *grpc.Server) {
		pb.RegisterUserServiceServer(s, &UserServiceImpl{})
		pb.RegisterOrderServiceServer(s, &OrderServiceImpl{})
		pb.RegisterProductServiceServer(s, &ProductServiceImpl{})
	})

	// 启动服务器
	logger.Info("starting multi-service server", "port", 9000)
	if err := server.Start(); err != nil {
		logger.Error("failed to start server", "error", err)
	}
}
