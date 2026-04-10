package trace

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// Gin instruments Gin requests with OpenTelemetry tracing.
func Gin(service string) gin.HandlerFunc {
	return otelgin.Middleware(service)
}
