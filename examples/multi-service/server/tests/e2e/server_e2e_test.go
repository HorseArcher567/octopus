package e2e

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/HorseArcher567/octopus/examples/multi-service/proto/pb"
	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/order"
	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/product"
	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/shared"
	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/user"
	"github.com/HorseArcher567/octopus/pkg/assemble"
	"github.com/HorseArcher567/octopus/pkg/config"
	"github.com/HorseArcher567/octopus/pkg/rpc"
	_ "github.com/go-sql-driver/mysql"
)

func TestServerE2E(t *testing.T) {
	dsn := os.Getenv("OCTOPUS_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set OCTOPUS_TEST_MYSQL_DSN to run e2e test")
	}

	rpcPort := freePort(t)
	apiPort := freePort(t)
	cfgPath := writeConfig(t, dsn, rpcPort, apiPort)
	prepareSchema(t, dsn)

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	a, err := assemble.New(
		cfg,
		assemble.WithSetupSteps(shared.SetupHello()),
		assemble.With(
			shared.AssembleHello,
			user.Assemble,
			order.Assemble,
			product.Assemble,
		),
	)
	if err != nil {
		t.Fatalf("build app: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- a.Run(ctx)
	}()

	if err := waitHTTPReady(apiPort, 8*time.Second); err != nil {
		t.Fatalf("api not ready: %v", err)
	}

	conn, err := rpc.NewClient(fmt.Sprintf("127.0.0.1:%d", rpcPort))
	if err != nil {
		t.Fatalf("new rpc client: %v", err)
	}
	defer conn.Close()

	userClient := pb.NewUserClient(conn)
	orderClient := pb.NewOrderClient(conn)
	productClient := pb.NewProductClient(conn)

	testCtx, testCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer testCancel()

	createUserResp, err := userClient.CreateUser(testCtx, &pb.CreateUserRequest{Username: "e2e_user", Email: "e2e_user@example.com"})
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if _, err := userClient.GetUser(testCtx, &pb.GetUserRequest{UserId: createUserResp.UserId}); err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	if _, err := orderClient.CreateOrder(testCtx, &pb.CreateOrderRequest{UserId: createUserResp.UserId, ProductName: "e2e-product", Amount: 19.99}); err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	if _, err := productClient.ListProducts(testCtx, &pb.ListProductsRequest{Page: 1, PageSize: 10}); err != nil {
		t.Fatalf("ListProducts failed: %v", err)
	}

	httpClient := &http.Client{Timeout: 3 * time.Second}

	httpCreateUserResp := struct {
		UserID int64 `json:"user_id"`
	}{}
	if err := doJSON(httpClient, http.MethodPost, fmt.Sprintf("http://127.0.0.1:%d/users", apiPort), map[string]any{
		"username": "http_user",
		"email":    "http_user@example.com",
	}, &httpCreateUserResp); err != nil {
		t.Fatalf("http CreateUser failed: %v", err)
	}
	if httpCreateUserResp.UserID == 0 {
		t.Fatal("expected http CreateUser to return user id")
	}

	if err := doJSON(httpClient, http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/users/%d", apiPort, httpCreateUserResp.UserID), nil, &struct {
		UserID int64 `json:"user_id"`
	}{}); err != nil {
		t.Fatalf("http GetUser failed: %v", err)
	}

	httpCreateOrderResp := struct {
		OrderID int64 `json:"order_id"`
	}{}
	if err := doJSON(httpClient, http.MethodPost, fmt.Sprintf("http://127.0.0.1:%d/orders", apiPort), map[string]any{
		"user_id":      httpCreateUserResp.UserID,
		"product_name": "http-product",
		"amount":       28.8,
	}, &httpCreateOrderResp); err != nil {
		t.Fatalf("http CreateOrder failed: %v", err)
	}
	if httpCreateOrderResp.OrderID == 0 {
		t.Fatal("expected http CreateOrder to return order id")
	}

	if err := doJSON(httpClient, http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/orders/%d", apiPort, httpCreateOrderResp.OrderID), nil, &struct {
		OrderID int64 `json:"order_id"`
	}{}); err != nil {
		t.Fatalf("http GetOrder failed: %v", err)
	}

	httpProductsResp := struct {
		Products []struct {
			ProductID int64 `json:"product_id"`
		} `json:"products"`
		Total int32 `json:"total"`
	}{}
	if err := doJSON(httpClient, http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/products?page=1&page_size=10", apiPort), nil, &httpProductsResp); err != nil {
		t.Fatalf("http ListProducts failed: %v", err)
	}
	if httpProductsResp.Total == 0 {
		t.Fatal("expected http ListProducts to return products")
	}

	cancel()
	select {
	case runErr := <-done:
		if runErr != nil {
			t.Fatalf("app exited with error: %v", runErr)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("app did not shutdown in time")
	}
}

func freePort(t *testing.T) int {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("alloc port: %v", err)
	}
	defer lis.Close()
	return lis.Addr().(*net.TCPAddr).Port
}

func writeConfig(t *testing.T, dsn string, rpcPort, apiPort int) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := fmt.Sprintf(`logger:
  - name: default
    level: info
    format: text
    output: stdout

app:
  logger: default

hello:
  message: hello from custom setup step

rpcServer:
  name: multi-service-e2e
  host: 127.0.0.1
  port: %d
  enableReflection: false

apiServer:
  name: multi-service-e2e
  host: 127.0.0.1
  port: %d
  mode: release

mysql:
  - name: primary
    dsn: %q
    driverName: mysql
`, rpcPort, apiPort, dsn)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

func waitHTTPReady(port int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	url := fmt.Sprintf("http://127.0.0.1:%d/hello", port)
	client := &http.Client{Timeout: 400 * time.Millisecond}

	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return nil
			}
		}
		time.Sleep(150 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for %s", url)
}

func doJSON(client *http.Client, method, url string, body any, target any) error {
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
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(payload))
	}
	if target == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

func prepareSchema(t *testing.T, dsn string) {
	t.Helper()

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("open mysql: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("ping mysql: %v", err)
	}

	schemaPath := schemaFilePath(t)
	payload, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}

	for _, stmt := range splitSQLStatements(string(payload)) {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("exec schema statement %q: %v", stmt, err)
		}
	}
}

func schemaFilePath(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current file path")
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "schema.sql")
}

func splitSQLStatements(payload string) []string {
	parts := strings.Split(payload, ";")
	statements := make([]string, 0, len(parts))
	for _, part := range parts {
		stmt := strings.TrimSpace(part)
		if stmt == "" {
			continue
		}
		statements = append(statements, stmt)
	}
	return statements
}
