package promsvc

import (
	"github.com/k8s-practice/octopus/pkg/service"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
)

var (
	singleton *Service
)

func init() {
	// auto register builder
	service.RegisterBuilder(&Builder{})
}

type Service struct {
	enabled bool
	name    string
	server  *http.Server
	address string
	path    string
}

func MustRegister(collectors ...prometheus.Collector) {
	prometheus.MustRegister(collectors...)
}
