package ginsvc

import (
	"github.com/gin-gonic/gin"
	"github.com/k8s-practice/octopus/pkg/log"
	"github.com/k8s-practice/octopus/pkg/prometheus/metrics"
	"github.com/k8s-practice/octopus/pkg/service"
	"net/http"
	"reflect"
)

var (
	defaultBuilder = &builder{}
)

func init() {
	// auto register builder
	service.RegisterBuilder(defaultBuilder)
}

func Router() gin.IRouter {
	return defaultBuilder.service.Router()
}

type Service struct {
	enabled bool
	name    string
	server  *http.Server
	address string

	metrics *metrics.GinServerMetrics
}

func (svc *Service) Router() gin.IRouter {
	if svc == nil {
		log.Panicf("%s uninitialized", reflect.TypeOf(svc))
		return nil
	}
	return svc.server.Handler.(*gin.Engine)
}
