package httpsvc

import (
	"github.com/k8s-practice/octopus/pkg/log"
	"github.com/k8s-practice/octopus/pkg/prometheus/metrics"
	"github.com/k8s-practice/octopus/pkg/service"
	"net/http"
	"reflect"
)

var (
	singleton *Service
)

func init() {
	// auto register builder
	service.RegisterBuilder(&Builder{})
}

type Service struct {
	name    string
	server  *http.Server
	address string

	metrics *metrics.HttpServerMetrics
}

func ServeMux() *http.ServeMux {
	if singleton != nil {
		return singleton.server.Handler.(*ServeMuxWrapper).ServeMux
	} else {
		log.Panicf("%s uninitialized", reflect.TypeOf(singleton))
		return nil
	}
}
