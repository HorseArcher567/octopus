package api

import "github.com/gin-gonic/gin"

// Option customizes HTTP Server behavior.
type Option func(s *Server)

// WithMiddleware appends custom Gin middleware after the default middleware stack.
func WithMiddleware(mw ...gin.HandlerFunc) Option {
	return func(s *Server) {
		s.extraMiddleware = append(s.extraMiddleware, mw...)
	}
}

// WithoutDefaultMiddleware disables the built-in middleware stack.
func WithoutDefaultMiddleware() Option {
	return func(s *Server) {
		s.defaultMiddleware = false
	}
}
