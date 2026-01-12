package api

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HorseArcher567/octopus/pkg/api/middleware"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	"github.com/gin-gonic/gin"
)

// Engine 是 gin.Engine 的类型别名，便于在其他包中引用而不直接依赖 gin。
type Engine = gin.Engine

// Server 封装 Gin HTTP 服务的生命周期。
type Server struct {
	config *ServerConfig

	engine     *gin.Engine
	httpServer *http.Server
	listener   net.Listener

	log *slog.Logger
}

// NewServer 创建 HTTP API 服务器。
// 从 context 中获取 logger，如果没有则使用 slog.Default()。
func NewServer(ctx context.Context, cfg *ServerConfig, opts ...Option) *Server {
	if cfg == nil {
		panic("api: server config is nil")
	}

	log := xlog.FromContext(ctx).With("component", "api.server", "appName", cfg.AppName)

	s := &Server{
		config: cfg,
		log:    log,
	}

	for _, opt := range opts {
		opt(s)
	}

	// 初始化 Gin Engine
	if s.engine == nil {
		mode := cfg.Mode
		if mode == "" {
			mode = gin.ReleaseMode
		}
		gin.SetMode(mode)

		engine := gin.New()
		// 使用自定义 Recovery 与 Logging 中间件。
		engine.Use(
			middleware.Recovery(),
			middleware.Logging(),
		)
		s.engine = engine
	}

	// 如果配置了 pprof，则挂载到 /debug/pprof
	if cfg.EnablePProf {
		s.registerPProf()
	}

	return s
}

// Engine 返回内部的 gin.Engine，便于注册路由和中间件。
func (s *Server) Engine() *gin.Engine {
	return s.engine
}

// Use 向 Engine 添加中间件。
func (s *Server) Use(middlewares ...gin.HandlerFunc) {
	s.engine.Use(middlewares...)
}

// Start 启动 HTTP 服务器并阻塞，直到收到退出信号并完成优雅关闭。
func (s *Server) Start() error {
	if s.config.Port <= 0 {
		return fmt.Errorf("api: invalid port %d", s.config.Port)
	}

	host := s.config.Host
	if host == "" {
		host = "0.0.0.0"
	}
	addr := fmt.Sprintf("%s:%d", host, s.config.Port)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("api: failed to listen on %s: %w", addr, err)
	}
	s.listener = lis

	server := &http.Server{
		Addr:         addr,
		Handler:      s.engine,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}
	s.httpServer = server

	s.log.Info("starting api server", "addr", addr)

	// 启动 HTTP Server
	go func() {
		if err := server.Serve(lis); err != nil && err != http.ErrServerClosed {
			s.log.Error("api server stopped", "error", err)
		}
	}()

	// 等待退出信号并优雅关闭
	s.waitForShutdown()

	return nil
}

// Shutdown 优雅关闭 HTTP 服务器。
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}

// registerPProf 将 pprof 路由挂载到 /debug/pprof。
func (s *Server) registerPProf() {
	g := s.engine.Group("/debug/pprof")
	{
		g.GET("/", gin.WrapF(pprof.Index))
		g.GET("/cmdline", gin.WrapF(pprof.Cmdline))
		g.GET("/profile", gin.WrapF(pprof.Profile))
		g.POST("/symbol", gin.WrapF(pprof.Symbol))
		g.GET("/symbol", gin.WrapF(pprof.Symbol))
		g.GET("/trace", gin.WrapF(pprof.Trace))
		g.GET("/allocs", gin.WrapH(pprof.Handler("allocs")))
		g.GET("/block", gin.WrapH(pprof.Handler("block")))
		g.GET("/goroutine", gin.WrapH(pprof.Handler("goroutine")))
		g.GET("/heap", gin.WrapH(pprof.Handler("heap")))
		g.GET("/mutex", gin.WrapH(pprof.Handler("mutex")))
		g.GET("/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
	}
}

// waitForShutdown 等待关闭信号并执行优雅关闭。
func (s *Server) waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan

	s.log.Info("shutting down api server gracefully")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		s.log.Error("failed to shutdown api server", "error", err)
		return
	}

	s.log.Info("api server shutdown complete")
}
