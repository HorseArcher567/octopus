package middleware

import (
	"net/http"

	"github.com/HorseArcher567/octopus/pkg/xlog"
	"github.com/gin-gonic/gin"
)

// Recovery returns a simple panic recovery middleware.
// It recovers from panics, logs the error, and returns HTTP 500.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log := xlog.FromContext(c.Request.Context())
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
