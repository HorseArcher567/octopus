package httpsvc

import (
	"github.com/k8s-practice/octopus/pkg/log"
	"github.com/k8s-practice/octopus/pkg/prometheus/metrics"
	"github.com/k8s-practice/octopus/pkg/service"
	"github.com/k8s-practice/octopus/pkg/service/config"
	"github.com/k8s-practice/octopus/pkg/service/promsvc"
	"github.com/k8s-practice/octopus/pkg/util/structure"
	"net/http"
)

type Builder struct {
}

type Config struct {
	Http struct {
		Enabled    bool              `json:"enabled,omitempty" yaml:"enabled,omitempty"`
		Name       string            `json:"name,omitempty" yaml:"name,omitempty"`
		Address    string            `json:"address,omitempty" yaml:"address,omitempty"`
		Prometheus config.Prometheus `json:"prometheus,omitempty" yaml:"prometheus,omitempty"`
	} `json:"http,omitempty" yaml:"http,omitempty"`
}

func (builder *Builder) Build(bootConfig map[interface{}]interface{}, tag string) service.Entry {
	var conf Config
	if err := structure.UnmarshalWithTag(bootConfig, &conf, tag); err != nil {
		log.Panicln(err)
		return nil
	}

	serveMuxWrapper := newServeMuxWrapper()
	singleton = &Service{
		enabled: conf.Http.Enabled,
		name:    conf.Http.Name,
		server: &http.Server{
			Handler: serveMuxWrapper,
		},
		address: conf.Http.Address,
	}
	if conf.Http.Prometheus.Server.Enabled {
		singleton.metrics = metrics.NewHttpServerMetrics(conf.Http.Prometheus.Server.Namespace,
			conf.Http.Prometheus.Server.Subsystem)
		promsvc.MustRegister(singleton.metrics)

		if conf.Http.Prometheus.Server.CountsHandlingTime {
			singleton.metrics.EnableCountsHandlingTime()
		}
		serveMuxWrapper.Use(singleton.metrics.Interceptor())
	}

	return singleton
}
