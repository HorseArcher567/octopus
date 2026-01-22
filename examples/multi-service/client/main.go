package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/HorseArcher567/octopus/examples/multi-service/proto/pb"
	"github.com/HorseArcher567/octopus/pkg/app"
	"github.com/HorseArcher567/octopus/pkg/config"
	"github.com/HorseArcher567/octopus/pkg/xlog"
)

// AppConfig 应用配置
type AppConfig struct {
	app.Framework // 嵌入框架配置
	// 注意：RpcClients 已经在 app.Framework 中定义，无需重复
}

func main() {
	// 1. 加载配置
	var cfg AppConfig
	config.MustUnmarshal("config.yaml", &cfg)

	// 2. 初始化框架
	app.Init(&cfg.Framework)

	// 3. 注册测试任务
	app.AddJob("test-services", runServiceTests)

	// 4. 启动应用（执行完测试后自动退出）
	app.Run()
}

// runServiceTests 执行所有服务测试
func runServiceTests(ctx context.Context, log *xlog.Logger) error {
	conn := app.MustNewRpcClient("etcd:///multi-service-demo")

	// 创建 gRPC 服务客户端
	userClient := pb.NewUserClient(conn)
	orderClient := pb.NewOrderClient(conn)
	productClient := pb.NewProductClient(conn)

	// 创建带超时的 context
	testCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 执行测试
	log.Info("=== Testing UserService ===")
	testUserService(testCtx, userClient, log)

	log.Info("=== Testing OrderService ===")
	testOrderService(testCtx, orderClient, log)

	log.Info("=== Testing ProductService ===")
	testProductService(testCtx, productClient, log)

	log.Info("=== Testing ApiServer (HTTP) ===")
	testApiServer(log)

	log.Info("✅ All tests completed successfully!")
	return nil
}

// testUserService tests the UserService gRPC methods.
func testUserService(ctx context.Context, client pb.UserClient, log *xlog.Logger) {
	// Test GetUser.
	getUserResp, err := client.GetUser(ctx, &pb.GetUserRequest{
		UserId: 1001,
	})
	if err != nil {
		log.Error("❌ GetUser failed", "error", err)
		return
	}
	log.Info("✅ GetUser",
		"user_id", getUserResp.UserId,
		"username", getUserResp.Username,
		"email", getUserResp.Email)

	// Test CreateUser.
	createUserResp, err := client.CreateUser(ctx, &pb.CreateUserRequest{
		Username: "newuser",
		Email:    "newuser@example.com",
	})
	if err != nil {
		log.Error("❌ CreateUser failed", "error", err)
		return
	}
	log.Info("✅ CreateUser",
		"user_id", createUserResp.UserId,
		"message", createUserResp.Message)
}

// testOrderService tests the OrderService gRPC methods.
func testOrderService(ctx context.Context, client pb.OrderClient, log *xlog.Logger) {
	// Test GetOrder.
	getOrderResp, err := client.GetOrder(ctx, &pb.GetOrderRequest{
		OrderId: 2001,
	})
	if err != nil {
		log.Error("❌ GetOrder failed", "error", err)
		return
	}
	log.Info("✅ GetOrder",
		"order_id", getOrderResp.OrderId,
		"product", getOrderResp.ProductName,
		"amount", getOrderResp.Amount,
		"status", getOrderResp.Status)

	// Test CreateOrder.
	createOrderResp, err := client.CreateOrder(ctx, &pb.CreateOrderRequest{
		UserId:      1001,
		ProductName: "New Product",
		Amount:      199.99,
	})
	if err != nil {
		log.Error("❌ CreateOrder failed", "error", err)
		return
	}
	log.Info("✅ CreateOrder",
		"order_id", createOrderResp.OrderId,
		"message", createOrderResp.Message)
}

// testProductService tests the ProductService gRPC methods.
func testProductService(ctx context.Context, client pb.ProductClient, log *xlog.Logger) {
	// Test GetProduct.
	getProductResp, err := client.GetProduct(ctx, &pb.GetProductRequest{
		ProductId: 1,
	})
	if err != nil {
		log.Error("❌ GetProduct failed", "error", err)
		return
	}
	log.Info("✅ GetProduct",
		"product_id", getProductResp.ProductId,
		"name", getProductResp.Name,
		"price", getProductResp.Price,
		"stock", getProductResp.Stock)

	// Test ListProducts.
	listProductsResp, err := client.ListProducts(ctx, &pb.ListProductsRequest{
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		log.Error("❌ ListProducts failed", "error", err)
		return
	}
	log.Info("✅ ListProducts", "total", listProductsResp.Total)
	for i, product := range listProductsResp.Products {
		log.Info("  Product",
			"index", i+1,
			"name", product.Name,
			"price", product.Price,
			"stock", product.Stock)
	}
}

// testApiServer tests the HTTP API server's /hello endpoint.
func testApiServer(log *xlog.Logger) {
	// Allow API address to be overridden via environment variable.
	// Defaults to localhost:8080.
	apiAddr := os.Getenv("API_SERVER_ADDR")
	if apiAddr == "" {
		apiAddr = "http://localhost:8080"
	}

	url := apiAddr + "/hello"
	resp, err := http.Get(url)
	if err != nil {
		log.Error("❌ ApiServer /hello request failed", "error", err)
		return
	}
	defer resp.Body.Close()

	log.Info("✅ ApiServer /hello", "status", resp.StatusCode, "url", url)
}
