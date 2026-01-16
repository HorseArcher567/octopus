package api

import (
	"context"
	"fmt"
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

// Engine is a type alias for gin.Engine.
type Engine = gin.Engine

// Router is a type alias for gin.IRouter.
type Router = gin.IRouter

// Server encapsulates the lifecycle of a Gin HTTP server.
type Server struct {
	log    *xlog.Logger
	config *ServerConfig

	engine     *gin.Engine
	httpServer *http.Server
}

// MustNewServer creates a new Server and panics if initialization fails.
func MustNewServer(log *xlog.Logger, cfg *ServerConfig, opts ...Option) *Server {
	server, err := NewServer(log, cfg, opts...)
	if err != nil {
		panic(err)
	}
	return server
}

// NewServer creates a new Server with the given configuration.
// Functional options can be used to customize the server behavior.
func NewServer(log *xlog.Logger, cfg *ServerConfig, opts ...Option) (*Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	s := &Server{
		log:    log,
		config: cfg,
	}

	for _, opt := range opts {
		opt(s)
	}

	gin.SetMode(cfg.Mode)
	s.engine = gin.New()
	s.engine.Use(
		middleware.Recovery(),
		middleware.Logging(),
	)

	// Mount pprof routes if enabled.
	if cfg.EnablePProf {
		s.registerPProf()
	}

	return s, nil
}

// Engine returns the underlying gin.Engine for route registration.
func (s *Server) Engine() *Engine {
	return s.engine
}

// Start starts the HTTP server and blocks until receiving a shutdown signal.
//
// The server is started in a goroutine. If ListenAndServe fails with an error
// other than http.ErrServerClosed (normal shutdown), it will panic.
// The method blocks until a termination signal (SIGTERM or SIGINT) is received.
func (s *Server) Start() {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.engine,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}
	s.log.Info("starting api server", "addr", addr)

	go func() {
		err := s.httpServer.ListenAndServe()
		if err == nil || err == http.ErrServerClosed {
			s.log.Info("api server closed")
		} else {
			s.log.Error("failed to start api server", "error", err)
			panic(err)
		}
	}()

	s.waitForShutdown()
}

// Shutdown gracefully shuts down the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}

// registerPProf mounts pprof routes to /debug/pprof.
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

// waitForShutdown waits for a shutdown signal and performs graceful shutdown.
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
