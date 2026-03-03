package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/HorseArcher567/octopus/examples/multi-service/proto/pb"
	"github.com/HorseArcher567/octopus/pkg/app"
	"github.com/HorseArcher567/octopus/pkg/config"
)

func main() {
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	target := flag.String("target", "etcd:///multi-service-demo", "RPC target")
	apiURL := flag.String("api", "http://127.0.0.1:8090/hello", "API URL")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		panic(err)
	}
	a := app.New(cfg)
	defer a.Logger().Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := runDemo(ctx, a, *target, *apiURL); err != nil {
		panic(err)
	}
}

func runDemo(ctx context.Context, a *app.App, target, apiURL string) error {
	log := a.Logger()

	conn, err := a.NewRPCClient(target)
	if err != nil {
		return fmt.Errorf("new rpc client: %w", err)
	}
	defer conn.Close()

	userClient := pb.NewUserClient(conn)
	orderClient := pb.NewOrderClient(conn)
	productClient := pb.NewProductClient(conn)

	username := "demo_user"
	email := "demo_user@example.com"

	createUserResp, err := userClient.CreateUser(ctx, &pb.CreateUserRequest{Username: username, Email: email})
	if err != nil && (strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "Error 1062")) {
		// 用户名或邮箱已存在时，为演示用途生成一个新的用户名和邮箱重试
		suffix := time.Now().Unix()
		username = fmt.Sprintf("demo_user_%d", suffix)
		email = fmt.Sprintf("demo_user_%d@example.com", suffix)
		createUserResp, err = userClient.CreateUser(ctx, &pb.CreateUserRequest{Username: username, Email: email})
	}
	if err != nil {
		return fmt.Errorf("CreateUser: %w", err)
	}
	log.Info("CreateUser ok", "user_id", createUserResp.UserId)

	if _, err := userClient.GetUser(ctx, &pb.GetUserRequest{UserId: createUserResp.UserId}); err != nil {
		return fmt.Errorf("GetUser: %w", err)
	}
	log.Info("GetUser ok", "user_id", createUserResp.UserId)

	if _, err := orderClient.CreateOrder(ctx, &pb.CreateOrderRequest{UserId: createUserResp.UserId, ProductName: "demo-product", Amount: 99.9}); err != nil {
		return fmt.Errorf("CreateOrder: %w", err)
	}
	log.Info("CreateOrder ok", "user_id", createUserResp.UserId)

	if _, err := productClient.ListProducts(ctx, &pb.ListProductsRequest{Page: 1, PageSize: 10}); err != nil {
		return fmt.Errorf("ListProducts: %w", err)
	}
	log.Info("ListProducts ok")

	if err := checkAPI(apiURL); err != nil {
		return fmt.Errorf("API check: %w", err)
	}
	log.Info("API check ok", "url", apiURL)
	return nil
}

func checkAPI(url string) error {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}
