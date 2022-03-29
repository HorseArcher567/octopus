package promsvc

import (
	"github.com/k8s-practice/octopus/pkg/log"
	"github.com/k8s-practice/octopus/pkg/service"
	"github.com/prometheus/client_golang/prometheus"
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

func MustRegister(collectors ...prometheus.Collector) {
	defaultBuilder.service.MustRegister(collectors...)
}

type Service struct {
	enabled bool
	name    string
	server  *http.Server
	address string
	path    string

	registerer prometheus.Registerer
}

func (svc *Service) MustRegister(collectors ...prometheus.Collector) {
	if svc == nil {
		log.Panicf("%s uninitialized", reflect.TypeOf(svc))
		return
	}
	svc.registerer.MustRegister(collectors...)
}
