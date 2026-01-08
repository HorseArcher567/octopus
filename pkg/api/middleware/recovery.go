package middleware

import (
	"net/http"

	"github.com/HorseArcher567/octopus/pkg/logger"
	"github.com/gin-gonic/gin"
)

// Recovery 是一个简单的 panic 恢复中间件。
// 发生 panic 时返回 500，并记录错误日志。
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log := logger.FromContext(c.Request.Context())
				log.Error("panic recovered in http handler", "panic", r)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    http.StatusInternalServerError,
					"message": "internal server error",
				})
			}
		}()

		c.Next()
	}
}
