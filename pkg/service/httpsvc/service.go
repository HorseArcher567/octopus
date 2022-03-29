package httpsvc

import (
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

func ServeMux() *http.ServeMux {
	return defaultBuilder.service.ServeMux()
}

type Service struct {
	enabled bool
	name    string
	server  *http.Server
	address string

	metrics *metrics.HttpServerMetrics
}

func (svc *Service) ServeMux() *http.ServeMux {
	if svc == nil {
		log.Panicf("%s uninitialized", reflect.TypeOf(svc))
		return nil
	}
	return svc.server.Handler.(*ServeMuxWrapper).ServeMux
}
