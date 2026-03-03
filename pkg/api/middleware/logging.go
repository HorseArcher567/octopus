package middleware

import (
	"time"

	"github.com/HorseArcher567/octopus/pkg/xlog"
	"github.com/gin-gonic/gin"
)

// LoggerInjector injects the given Logger into the Context of each request.
// Subsequent middlewares and handlers can retrieve the same logger via xlog.Get(c.Request.Context()).
func LoggerInjector(base *xlog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Put the base logger into the request context.
		ctx := xlog.Put(c.Request.Context(), base)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// Logging returns a simple HTTP request logging middleware.
// It logs method, path, status, latency and client IP for each request.
func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process the request.
		c.Next()

		latency := time.Since(start)

		log := xlog.Get(c.Request.Context())
		log.Info("http request",
			"method", c.Request.Method,
			"path", c.FullPath(),
			"status", c.Writer.Status(),
			"latency", latency.String(),
			"client_ip", c.ClientIP(),
		)
	}
}
