package promsvc

import (
	"github.com/k8s-practice/octopus/pkg/log"
	"github.com/k8s-practice/octopus/pkg/service"
	"github.com/k8s-practice/octopus/pkg/util/structure"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

const (
	defaultMetricsPath = "/metrics"
)

type Builder struct {
}

type Config struct {
	Prometheus struct {
		Enabled bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
		Name    string `json:"name,omitempty" yaml:"name,omitempty"`
		Address string `json:"address,omitempty" yaml:"address,omitempty"`
		Path    string `json:"path,omitempty" yaml:"path,omitempty"`
	} `json:"prometheus,omitempty" yaml:"prometheus,omitempty"`
}

func (builder *Builder) Build(bootConfig map[interface{}]interface{}, tag string) service.Entry {
	var conf Config
	if err := structure.UnmarshalWithTag(bootConfig, &conf, tag); err != nil {
		log.Panicln(err)
		return nil
	}

	if len(conf.Prometheus.Path) == 0 {
		conf.Prometheus.Path = defaultMetricsPath
	}
	mux := http.NewServeMux()
	mux.Handle(conf.Prometheus.Path, promhttp.Handler())
	singleton = &Service{
		enabled: conf.Prometheus.Enabled,
		name:    conf.Prometheus.Name,
		server: &http.Server{
			Handler: mux,
		},
		address: conf.Prometheus.Address,
		path:    conf.Prometheus.Path,
	}

	return singleton
}
