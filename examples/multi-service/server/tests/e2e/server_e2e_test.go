package e2e

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/HorseArcher567/octopus/examples/multi-service/proto/pb"
	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/bootstrap"
	"github.com/HorseArcher567/octopus/pkg/app"
	"github.com/HorseArcher567/octopus/pkg/config"
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

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	a := app.New(cfg)
	infra := bootstrap.NewInfraModule()
	a.Use(
		infra,
		bootstrap.NewRPCModule(infra),
		bootstrap.NewAPIModule(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- a.Run(ctx)
	}()

	if err := waitHTTPReady(apiPort, 8*time.Second); err != nil {
		t.Fatalf("api not ready: %v", err)
	}

	conn, err := a.NewRPCClient(fmt.Sprintf("127.0.0.1:%d", rpcPort))
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
  level: info
  format: text
  output: stdout

rpcServer:
  name: multi-service-e2e
  host: 127.0.0.1
  port: %d
  enableReflection: false
  advertiseAddr: ""

apiServer:
  name: multi-service-e2e
  host: 127.0.0.1
  port: %d
  mode: release

resources:
  mysql:
    primary:
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
