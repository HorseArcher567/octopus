package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler exposes the health registry over HTTP.
func Handler(reg *Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		report := reg.Check(c.Request.Context())
		if report.Status == StatusDown {
			c.JSON(http.StatusServiceUnavailable, report)
			return
		}
		c.JSON(http.StatusOK, report)
	}
}
