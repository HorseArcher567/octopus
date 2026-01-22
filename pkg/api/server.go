package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"

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
func MustNewServer(log *xlog.Logger, config *ServerConfig, opts ...Option) *Server {
	server, err := NewServer(log, config, opts...)
	if err != nil {
		panic(err)
	}
	return server
}

// NewServer creates a new Server with the given configuration.
// Functional options can be used to customize the server behavior.
func NewServer(log *xlog.Logger, config *ServerConfig, opts ...Option) (*Server, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	s := &Server{
		log:    log,
		config: config,
	}

	for _, opt := range opts {
		opt(s)
	}

	gin.SetMode(config.Mode)
	s.engine = gin.New()
	s.engine.Use(
		middleware.Recovery(),
		middleware.Logging(),
	)

	// Mount pprof routes if enabled.
	if config.EnablePProf {
		s.registerPProf()
	}

	return s, nil
}

// Engine returns the underlying gin.Engine for route registration.
func (s *Server) Engine() *Engine {
	return s.engine
}

// Start starts the HTTP server in a background goroutine and returns immediately.
// If the server fails to start, it will panic (to fail-fast during startup).
// Use Stop to gracefully shut down the server.
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.engine,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}

	s.log.Info("starting api server", "addr", addr)

	// Start server in background
	go func() {
		err := s.httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			s.log.Error("api server stopped unexpectedly", "error", err)
			panic(err) // Fail fast if server crashes unexpectedly
		}
	}()

	return nil
}

// Stop gracefully shuts down the HTTP server with the given context for timeout control.
func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}

	s.log.Info("shutting down api server gracefully")
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.log.Error("failed to shutdown api server", "error", err)
		return err
	}

	s.log.Info("api server shutdown complete")
	return nil
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
