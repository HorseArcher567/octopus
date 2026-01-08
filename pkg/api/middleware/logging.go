package middleware

import (
	"time"

	"github.com/HorseArcher567/octopus/pkg/logger"
	"github.com/gin-gonic/gin"
)

// Logging 返回一个简单的 HTTP 请求日志中间件。
// 会记录 method、path、status、latency 等信息。
func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 处理请求
		c.Next()

		latency := time.Since(start)

		log := logger.FromContext(c.Request.Context())
		log.Info("http request",
			"method", c.Request.Method,
			"path", c.FullPath(),
			"status", c.Writer.Status(),
			"latency", latency.String(),
			"client_ip", c.ClientIP(),
		)
	}
}
