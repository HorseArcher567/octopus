package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type checker struct {
	name string
	err  error
}

func (c checker) Name() string                { return c.name }
func (c checker) Check(context.Context) error { return c.err }

func TestRegistryCheck(t *testing.T) {
	reg := New()
	_ = reg.Register(checker{name: "db"})
	_ = reg.Register(checker{name: "redis", err: errors.New("down")})

	report := reg.Check(context.Background())
	if report.Status != StatusDown {
		t.Fatalf("expected DOWN, got %s", report.Status)
	}
	if report.Details["db"].Status != StatusUp {
		t.Fatalf("expected db UP, got %+v", report.Details["db"])
	}
	if report.Details["redis"].Status != StatusDown {
		t.Fatalf("expected redis DOWN, got %+v", report.Details["redis"])
	}
}

func TestHandlerStatusCode(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	reg := New()
	_ = reg.Register(checker{name: "db", err: errors.New("down")})

	r := gin.New()
	r.GET("/health", Handler(reg))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}
