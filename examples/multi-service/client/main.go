package main

import (
	"bytes"
	"context"
	"encoding/json"
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
	a, err := app.FromConfig(cfg)
	if err != nil {
		panic(err)
	}
	defer a.Logger().Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := runDemo(ctx, a, *target, *apiURL); err != nil {
		panic(err)
	}
}

func runDemo(ctx context.Context, a *app.App, target, apiURL string) error {
	log := a.Logger()

	userID, err := runRPCDemo(ctx, a, target)
	if err != nil {
		return err
	}
	log.Info("RPC demo ok", "user_id", userID)

	if err := runHTTPDemo(apiURL); err != nil {
		return err
	}
	log.Info("HTTP demo ok", "api_url", apiURL)
	return nil
}

func runRPCDemo(ctx context.Context, a *app.App, target string) (int64, error) {
	log := a.Logger()

	conn, err := a.NewRPCClient(target)
	if err != nil {
		return 0, fmt.Errorf("new rpc client: %w", err)
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
		return 0, fmt.Errorf("CreateUser: %w", err)
	}
	log.Info("CreateUser ok", "user_id", createUserResp.UserId)

	if _, err := userClient.GetUser(ctx, &pb.GetUserRequest{UserId: createUserResp.UserId}); err != nil {
		return 0, fmt.Errorf("GetUser: %w", err)
	}
	log.Info("GetUser ok", "user_id", createUserResp.UserId)

	if _, err := orderClient.CreateOrder(ctx, &pb.CreateOrderRequest{UserId: createUserResp.UserId, ProductName: "demo-product", Amount: 99.9}); err != nil {
		return 0, fmt.Errorf("CreateOrder: %w", err)
	}
	log.Info("CreateOrder ok", "user_id", createUserResp.UserId)

	if _, err := productClient.ListProducts(ctx, &pb.ListProductsRequest{Page: 1, PageSize: 10}); err != nil {
		return 0, fmt.Errorf("ListProducts: %w", err)
	}
	log.Info("ListProducts ok")
	return createUserResp.UserId, nil
}

func runHTTPDemo(apiURL string) error {
	baseURL, err := trimHelloURL(apiURL)
	if err != nil {
		return fmt.Errorf("invalid api url: %w", err)
	}

	if err := checkAPI(apiURL); err != nil {
		return fmt.Errorf("API check: %w", err)
	}

	createUserResp := struct {
		UserID int64 `json:"user_id"`
	}{}
	suffix := time.Now().Unix()
	if err := doJSON(http.MethodPost, baseURL+"/users", map[string]any{
		"username": fmt.Sprintf("http_demo_user_%d", suffix),
		"email":    fmt.Sprintf("http_demo_user_%d@example.com", suffix),
	}, &createUserResp); err != nil {
		return fmt.Errorf("http CreateUser: %w", err)
	}

	if err := doJSON(http.MethodGet, fmt.Sprintf("%s/users/%d", baseURL, createUserResp.UserID), nil, &struct {
		UserID int64 `json:"user_id"`
	}{}); err != nil {
		return fmt.Errorf("http GetUser: %w", err)
	}

	if err := doJSON(http.MethodPost, baseURL+"/orders", map[string]any{
		"user_id":      createUserResp.UserID,
		"product_name": "http-demo-product",
		"amount":       88.8,
	}, &struct {
		OrderID int64 `json:"order_id"`
	}{}); err != nil {
		return fmt.Errorf("http CreateOrder: %w", err)
	}

	if err := doJSON(http.MethodGet, baseURL+"/products?page=1&page_size=10", nil, &struct {
		Total int32 `json:"total"`
	}{}); err != nil {
		return fmt.Errorf("http ListProducts: %w", err)
	}
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

func doJSON(method, url string, body any, target any) error {
	client := &http.Client{Timeout: 5 * time.Second}

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}
	if target == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

func trimHelloURL(apiURL string) (string, error) {
	trimmed := strings.TrimRight(apiURL, "/")
	if trimmed == "" {
		return "", fmt.Errorf("empty api url")
	}
	if strings.HasSuffix(trimmed, "/hello") {
		return strings.TrimSuffix(trimmed, "/hello"), nil
	}
	return trimmed, nil
}
