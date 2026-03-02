package main

import (
	"context"
	"flag"

	_ "github.com/go-sql-driver/mysql" // MySQL 驱动

	"github.com/HorseArcher567/octopus/examples/multi-service/proto/pb"
	"github.com/HorseArcher567/octopus/examples/multi-service/server/models"
	"github.com/HorseArcher567/octopus/examples/multi-service/server/repository"
	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/HorseArcher567/octopus/pkg/app"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserServer 用户服务实现
type UserServer struct {
	BaseServer
	pb.UnimplementedUserServer
	userRepo repository.UserRepository
}

func (s *UserServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	log := s.Log(ctx, "user_id", req.UserId)
	log.Info("get user called")

	user, err := s.userRepo.GetByID(ctx, req.UserId)
	if err != nil {
		log.Error("failed to get user", "error", err)
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &pb.GetUserResponse{
		UserId:   user.ID,
		Username: user.Username,
		Email:    user.Email,
	}, nil
}

func (s *UserServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	log := s.Log(ctx, "username", req.Username, "email", req.Email)
	log.Info("create user called")

	userID, err := s.userRepo.Create(ctx, &models.User{
		Username: req.Username,
		Email:    req.Email,
	})
	if err != nil {
		log.Error("failed to create user", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &pb.CreateUserResponse{
		UserId:  userID,
		Message: "User created successfully",
	}, nil
}

// OrderServer 订单服务实现
type OrderServer struct {
	BaseServer
	pb.UnimplementedOrderServer
	orderRepo repository.OrderRepository
}

func (s *OrderServer) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	log := s.Log(ctx, "order_id", req.OrderId)
	log.Info("get order called")

	order, err := s.orderRepo.GetByID(ctx, req.OrderId)
	if err != nil {
		log.Error("failed to get order", "error", err)
		return nil, status.Errorf(codes.NotFound, "order not found: %v", err)
	}

	return &pb.GetOrderResponse{
		OrderId:     order.OrderID,
		UserId:      order.UserID,
		ProductName: order.ProductName,
		Amount:      order.Amount,
		Status:      order.Status,
	}, nil
}

func (s *OrderServer) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	log := s.Log(ctx, "user_id", req.UserId, "product", req.ProductName, "amount", req.Amount)
	log.Info("create order called")

	orderID, err := s.orderRepo.Create(ctx, &models.Order{
		UserID:      req.UserId,
		ProductName: req.ProductName,
		Amount:      req.Amount,
		Status:      "pending",
	})
	if err != nil {
		log.Error("failed to create order", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to create order: %v", err)
	}

	return &pb.CreateOrderResponse{
		OrderId: orderID,
		Message: "Order created successfully",
	}, nil
}

// ProductServer 产品服务实现
type ProductServer struct {
	BaseServer
	pb.UnimplementedProductServer
	productRepo repository.ProductRepository
}

func (s *ProductServer) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	log := s.Log(ctx, "product_id", req.ProductId)
	log.Info("get product called")

	product, err := s.productRepo.GetByID(ctx, req.ProductId)
	if err != nil {
		log.Error("failed to get product", "error", err)
		return nil, status.Errorf(codes.NotFound, "product not found: %v", err)
	}

	return &pb.GetProductResponse{
		ProductId:   product.ProductID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       int32(product.Stock),
	}, nil
}

func (s *ProductServer) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	log := s.Log(ctx, "page", req.Page, "page_size", req.PageSize)
	log.Info("list products called")

	products, total, err := s.productRepo.List(ctx, req.Page, req.PageSize)
	if err != nil {
		log.Error("failed to list products", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to list products: %v", err)
	}

	// 转换为 protobuf 响应
	pbProducts := make([]*pb.GetProductResponse, 0, len(products))
	for _, p := range products {
		pbProducts = append(pbProducts, &pb.GetProductResponse{
			ProductId:   p.ProductID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Stock:       int32(p.Stock),
		})
	}

	return &pb.ListProductsResponse{
		Products: pbProducts,
		Total:    int32(total),
	}, nil
}

const mysqlPrimary = "primary"

func main() {
	configFile := flag.String("config", "config.yaml", "配置文件路径 (默认: config.yaml)")
	flag.Parse()

	app.MustRun(*configFile, wire)
}

func wire() error {
	db, err := app.MySQL(mysqlPrimary)
	if err != nil {
		return err
	}
	if err := initDatabase(db); err != nil {
		return err
	}

	logger := app.Logger()
	userRepo := repository.NewUserRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	productRepo := repository.NewProductRepository(db)

	registerRPC(userRepo, orderRepo, productRepo, logger)
	registerAPI()
	return nil
}

func registerRPC(
	userRepo repository.UserRepository,
	orderRepo repository.OrderRepository,
	productRepo repository.ProductRepository,
	logger *xlog.Logger,
) {
	app.RegisterRpcServices(func(s *grpc.Server) {
		pb.RegisterUserServer(s, &UserServer{
			BaseServer: BaseServer{logger: logger},
			userRepo:   userRepo,
		})
		pb.RegisterOrderServer(s, &OrderServer{
			BaseServer: BaseServer{logger: logger},
			orderRepo:  orderRepo,
		})
		pb.RegisterProductServer(s, &ProductServer{
			BaseServer:  BaseServer{logger: logger},
			productRepo: productRepo,
		})
	})
}

func registerAPI() {
	app.RegisterApiRoutes(func(engine *api.Engine) {
		engine.GET("/hello", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "hello from apiServer",
			})
		})
	})
}
