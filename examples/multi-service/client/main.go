package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/HorseArcher567/octopus/examples/multi-service/proto/pb"
	"github.com/HorseArcher567/octopus/pkg/config"
	"github.com/HorseArcher567/octopus/pkg/rpc"
)

func main() {
	ctx := context.Background()

	// 方式1：从配置文件加载（推荐）
	// 加载配置文件
	cfg, err := config.LoadWithEnv("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 从配置文件中创建客户端
	conn, err := rpc.NewClientFromConfig(ctx, cfg, "client")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// 方式2：直接使用代码配置（备选）
	// conn, err := rpc.NewClient(ctx, &rpc.ClientConfig{
	// 	AppName:  "multi-service-demo",
	// 	EtcdAddr: []string{"localhost:2379"},
	// })
	// if err != nil {
	// 	log.Fatalf("Failed to connect: %v", err)
	// }
	// defer conn.Close()

	// 方式3：直连模式（用于开发测试）
	// conn, err := rpc.NewClient(ctx, &rpc.ClientConfig{
	// 	Endpoints: []string{"localhost:9000", "localhost:9001"},
	// })
	// if err != nil {
	// 	log.Fatalf("Failed to connect: %v", err)
	// }
	// defer conn.Close()

	// 创建不同的 gRPC 服务客户端（共享同一个连接）
	userClient := pb.NewUserClient(conn)
	orderClient := pb.NewOrderClient(conn)
	productClient := pb.NewProductClient(conn)

	// 创建带超时的 context，用于所有 RPC 调用
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("=== Testing UserService ===")
	testUserService(ctx, userClient)

	log.Println("\n=== Testing OrderService ===")
	testOrderService(ctx, orderClient)

	log.Println("\n=== Testing ProductService ===")
	testProductService(ctx, productClient)

	log.Println("\n=== Testing ApiServer (HTTP) ===")
	testApiServer()

	log.Println("\n✅ All tests completed successfully!")
}

func testUserService(ctx context.Context, client pb.UserClient) {
	// 测试 GetUser
	getUserResp, err := client.GetUser(ctx, &pb.GetUserRequest{
		UserId: 1001,
	})
	if err != nil {
		log.Printf("❌ GetUser failed: %v", err)
		return
	}
	log.Printf("✅ GetUser: user_id=%d, username=%s, email=%s",
		getUserResp.UserId, getUserResp.Username, getUserResp.Email)

	// 测试 CreateUser
	createUserResp, err := client.CreateUser(ctx, &pb.CreateUserRequest{
		Username: "newuser",
		Email:    "newuser@example.com",
	})
	if err != nil {
		log.Printf("❌ CreateUser failed: %v", err)
		return
	}
	log.Printf("✅ CreateUser: user_id=%d, message=%s",
		createUserResp.UserId, createUserResp.Message)
}

func testOrderService(ctx context.Context, client pb.OrderClient) {
	// 测试 GetOrder
	getOrderResp, err := client.GetOrder(ctx, &pb.GetOrderRequest{
		OrderId: 2001,
	})
	if err != nil {
		log.Printf("❌ GetOrder failed: %v", err)
		return
	}
	log.Printf("✅ GetOrder: order_id=%d, product=%s, amount=%.2f, status=%s",
		getOrderResp.OrderId, getOrderResp.ProductName, getOrderResp.Amount, getOrderResp.Status)

	// 测试 CreateOrder
	createOrderResp, err := client.CreateOrder(ctx, &pb.CreateOrderRequest{
		UserId:      1001,
		ProductName: "New Product",
		Amount:      199.99,
	})
	if err != nil {
		log.Printf("❌ CreateOrder failed: %v", err)
		return
	}
	log.Printf("✅ CreateOrder: order_id=%d, message=%s",
		createOrderResp.OrderId, createOrderResp.Message)
}

func testProductService(ctx context.Context, client pb.ProductClient) {
	// 测试 GetProduct
	getProductResp, err := client.GetProduct(ctx, &pb.GetProductRequest{
		ProductId: 1,
	})
	if err != nil {
		log.Printf("❌ GetProduct failed: %v", err)
		return
	}
	log.Printf("✅ GetProduct: product_id=%d, name=%s, price=%.2f, stock=%d",
		getProductResp.ProductId, getProductResp.Name, getProductResp.Price, getProductResp.Stock)

	// 测试 ListProducts
	listProductsResp, err := client.ListProducts(ctx, &pb.ListProductsRequest{
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		log.Printf("❌ ListProducts failed: %v", err)
		return
	}
	log.Printf("✅ ListProducts: total=%d products", listProductsResp.Total)
	for i, product := range listProductsResp.Products {
		log.Printf("   [%d] %s - $%.2f (stock: %d)",
			i+1, product.Name, product.Price, product.Stock)
	}
}

// testApiServer 测试 HTTP ApiServer 的 /hello 接口。
func testApiServer() {
	// 允许通过环境变量覆盖 API 地址，默认为本机 8080 端口。
	apiAddr := os.Getenv("API_SERVER_ADDR")
	if apiAddr == "" {
		apiAddr = "http://localhost:8080"
	}

	url := apiAddr + "/hello"
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("❌ ApiServer /hello request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	log.Printf("✅ ApiServer /hello: status=%d, url=%s", resp.StatusCode, url)
}
