package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/HorseArcher567/octopus/pkg/xlog"
	"github.com/gin-gonic/gin"
)

func TestServerRegisterRunAndStop(t *testing.T) {
	log := xlog.MustNew(nil)
	defer log.Close()

	port := freePort(t)
	server, err := NewServer(log, &ServerConfig{
		Name: "api-test",
		Host: "127.0.0.1",
		Port: port,
		Mode: "release",
	})
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	if err := server.Register(func(engine *Engine) {
		engine.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})
	}); err != nil {
		t.Fatalf("register routes: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- server.Run(ctx)
	}()

	waitHTTPReady(t, fmt.Sprintf("http://127.0.0.1:%d/health", port), 5*time.Second)

	cancel()
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer stopCancel()
	if err := server.Stop(stopCtx); err != nil {
		t.Fatalf("stop server: %v", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("run returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not stop in time")
	}
}

func freePort(t *testing.T) int {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer lis.Close()
	return lis.Addr().(*net.TCPAddr).Port
}

func waitHTTPReady(t *testing.T, url string, timeout time.Duration) {
	t.Helper()
	client := &http.Client{Timeout: 200 * time.Millisecond}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for %s", url)
}
