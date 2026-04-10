package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGinMiddlewareAndHandler(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(GinMiddleware())
	r.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "pong") })
	r.GET("/metrics", gin.WrapH(Handler()))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	mw := httptest.NewRecorder()
	mreq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	r.ServeHTTP(mw, mreq)
	if mw.Code != http.StatusOK {
		t.Fatalf("expected metrics 200, got %d", mw.Code)
	}
	if body := mw.Body.String(); body == "" {
		t.Fatal("expected metrics body to be non-empty")
	}
}
