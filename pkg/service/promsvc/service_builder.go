package promsvc

import (
	"github.com/k8s-practice/octopus/pkg/log"
	"github.com/k8s-practice/octopus/pkg/service"
	"github.com/k8s-practice/octopus/pkg/util/structure"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"sync"
)

const (
	defaultMetricsPath = "/metrics"
)

type builder struct {
	sync.Once
	service *Service
}

type Config struct {
	Prometheus struct {
		Enabled bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
		Name    string `json:"name,omitempty" yaml:"name,omitempty"`
		Address string `json:"address,omitempty" yaml:"address,omitempty"`
		Path    string `json:"path,omitempty" yaml:"path,omitempty"`
	} `json:"prometheus,omitempty" yaml:"prometheus,omitempty"`
}

func (b *builder) Build(bootConfig map[interface{}]interface{}, tag string) service.Entry {
	b.Do(func() {
		var conf Config
		if err := structure.UnmarshalWithTag(bootConfig, &conf, tag); err != nil {
			log.Panicln(err)
			return
		}

		if len(conf.Prometheus.Path) == 0 {
			conf.Prometheus.Path = defaultMetricsPath
		}
		mux := http.NewServeMux()
		svc := &Service{
			enabled: conf.Prometheus.Enabled,
			name:    conf.Prometheus.Name,
			server: &http.Server{
				Handler: mux,
			},
			address:  conf.Prometheus.Address,
			path:     conf.Prometheus.Path,
			registry: prometheus.NewRegistry(),
		}
		mux.Handle(conf.Prometheus.Path, promhttp.HandlerFor(svc.registry, promhttp.HandlerOpts{}))

		b.service = svc
	})

	return b.service
}
