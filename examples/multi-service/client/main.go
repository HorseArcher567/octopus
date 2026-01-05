package main

import (
	"context"
	"log"
	"time"

	"github.com/HorseArcher567/octopus/examples/multi-service/proto/pb"
	"github.com/HorseArcher567/octopus/pkg/rpc"
)

func main() {
	// 通过 etcd 服务发现连接到服务器
	// AppName 是应用注册名，一个连接可以访问其下的所有 gRPC 服务
	conn, err := rpc.NewClient(context.Background(), &rpc.ClientConfig{
		AppName:  "multi-service-demo",
		EtcdAddr: []string{"localhost:2379"},
	})
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

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
