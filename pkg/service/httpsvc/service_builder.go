package httpsvc

import (
	"github.com/k8s-practice/octopus/pkg/log"
	"github.com/k8s-practice/octopus/pkg/prometheus/metrics"
	"github.com/k8s-practice/octopus/pkg/service"
	"github.com/k8s-practice/octopus/pkg/service/config"
	"github.com/k8s-practice/octopus/pkg/service/promsvc"
	"github.com/k8s-practice/octopus/pkg/util/structure"
	"net/http"
	"sync"
)

type Config struct {
	Http struct {
		Enabled              bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
		Name                 string `json:"name,omitempty" yaml:"name,omitempty"`
		Address              string `json:"address,omitempty" yaml:"address,omitempty"`
		config.Prometheus    `json:"prometheus,omitempty" yaml:"prometheus,omitempty"`
		config.OpenTelemetry `json:"openTelemetry,omitempty" yaml:"openTelemetry,omitempty"`
	} `json:"http,omitempty" yaml:"http,omitempty"`
}

type builder struct {
	sync.Once
	service *Service
}

func (b *builder) Build(bootConfig map[interface{}]interface{}, tag string) service.Entry {
	b.Do(func() {
		var conf Config
		if err := structure.UnmarshalWithTag(bootConfig, &conf, tag); err != nil {
			log.Panicln(err)
			return
		}

		serveMuxWrapper := newServeMuxWrapper()
		svc := &Service{
			enabled: conf.Http.Enabled,
			name:    conf.Http.Name,
			server: &http.Server{
				Handler: serveMuxWrapper,
			},
			address: conf.Http.Address,
		}

		if conf.Http.OpenTelemetry.Enabled {
			serveMuxWrapper.EnableTrace()
		}

		if conf.Http.Prometheus.Server.Enabled {
			svc.metrics = metrics.NewHttpServerMetrics(conf.Http.Prometheus.Server.Namespace,
				conf.Http.Prometheus.Server.Subsystem)
			promsvc.MustRegister(svc.metrics)

			if conf.Http.Prometheus.Server.CountsHandlingTime {
				svc.metrics.EnableCountsHandlingTime()
			}
			serveMuxWrapper.Use(svc.metrics.Interceptor())
		}

		b.service = svc
	})

	return b.service
}
