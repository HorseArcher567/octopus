package api

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

// Option 用于自定义 HTTP Server 的行为。
type Option func(s *Server)

// WithLogger 使用已有的 logger 实例。
func WithLogger(log *slog.Logger) Option {
	return func(s *Server) {
		if log != nil {
			s.log = log
		}
	}
}

// WithEngine 使用外部构造好的 gin.Engine。
// 如果不设置，默认使用 gin.New() 并由 Server 初始化常用中间件。
func WithEngine(engine *gin.Engine) Option {
	return func(s *Server) {
		if engine != nil {
			s.engine = engine
		}
	}
}
