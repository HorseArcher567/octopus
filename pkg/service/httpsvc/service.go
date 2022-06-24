package httpsvc

import (
	"github.com/k8s-practice/octopus/pkg/prometheus/metrics"
	"github.com/k8s-practice/octopus/pkg/service"
	"net/http"
)

var (
	defaultBuilder = &builder{}
)

func init() {
	// auto register builder
	service.RegisterBuilder(defaultBuilder)
}

func Handle(pattern string, handler http.Handler) {
	defaultBuilder.service.Handle(pattern, handler)
}

func HandleFunc(pattern string, handler func(w http.ResponseWriter, r *http.Request)) {
	defaultBuilder.service.HandleFunc(pattern, handler)
}

func Handler(r *http.Request) (h http.Handler, pattern string) {
	return defaultBuilder.service.Handler(r)
}

type Service struct {
	enabled bool
	name    string
	server  *http.Server
	address string

	metrics *metrics.HttpServerMetrics
}

func (svc *Service) Handle(pattern string, handler http.Handler) {
	svc.server.Handler.(*ServeMuxWrapper).Handle(pattern, handler)
}

func (svc *Service) HandleFunc(pattern string, handler func(w http.ResponseWriter, r *http.Request)) {
	svc.server.Handler.(*ServeMuxWrapper).HandleFunc(pattern, handler)
}

func (svc *Service) Handler(r *http.Request) (h http.Handler, pattern string) {
	return svc.server.Handler.(*ServeMuxWrapper).Handler(r)
}
