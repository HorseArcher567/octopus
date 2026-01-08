package main

import (
	"context"
	"flag"
	"log/slog"

	"github.com/HorseArcher567/octopus/examples/multi-service/proto/pb"
	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/HorseArcher567/octopus/pkg/app"
	"github.com/HorseArcher567/octopus/pkg/logger"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

// UserServer 用户服务实现
type UserServer struct {
	pb.UnimplementedUserServer
}

func (s *UserServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	log := logger.FromContext(ctx) // 获取带有 method, request_id 的 logger
	log.Info("get user called", "user_id", req.UserId)
	return &pb.GetUserResponse{
		UserId:   req.UserId,
		Username: "testuser",
		Email:    "test@example.com",
	}, nil
}

func (s *UserServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	log := logger.FromContext(ctx)
	log.Info("create user called",
		"username", req.Username,
		"email", req.Email,
	)
	return &pb.CreateUserResponse{
		UserId:  1001,
		Message: "User created successfully",
	}, nil
}

// OrderServer 订单服务实现
type OrderServer struct {
	pb.UnimplementedOrderServer
}

func (s *OrderServer) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	log := logger.FromContext(ctx)
	log.Info("get order called", "order_id", req.OrderId)
	return &pb.GetOrderResponse{
		OrderId:     req.OrderId,
		UserId:      1001,
		ProductName: "Sample Product",
		Amount:      99.99,
		Status:      "completed",
	}, nil
}

func (s *OrderServer) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	log := logger.FromContext(ctx)
	log.Info("create order called",
		"user_id", req.UserId,
		"product", req.ProductName,
		"amount", req.Amount,
	)
	return &pb.CreateOrderResponse{
		OrderId: 2001,
		Message: "Order created successfully",
	}, nil
}

// ProductServer 产品服务实现
type ProductServer struct {
	pb.UnimplementedProductServer
}

func (s *ProductServer) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	log := logger.FromContext(ctx)
	log.Info("get product called", "product_id", req.ProductId)
	return &pb.GetProductResponse{
		ProductId:   req.ProductId,
		Name:        "Sample Product",
		Description: "This is a sample product",
		Price:       99.99,
		Stock:       100,
	}, nil
}

func (s *ProductServer) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	log := logger.FromContext(ctx)
	log.Info("list products called",
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
	// 解析命令行参数
	configFile := flag.String("config", "config.yaml", "配置文件路径 (默认: config.yaml)")
	flag.Parse()

	// 使用 app 包的默认实例，封装配置加载、日志初始化和 RPC 服务器创建。
	app.Init(app.WithConfigFile(*configFile))

	// 注册多个服务
	app.RegisterRpcService(func(s *grpc.Server) {
		pb.RegisterUserServer(s, &UserServer{})
		pb.RegisterOrderServer(s, &OrderServer{})
		pb.RegisterProductServer(s, &ProductServer{})
	})

	// 注册一个简单的 HTTP Hello API
	app.RegisterApiRoutes(func(engine *api.Engine) {
		engine.GET("/hello", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "hello from apiServer",
			})
		})
	})

	// 启动应用
	if err := app.Run(); err != nil {
		slog.Error("failed to run app", "error", err)
	}
}
