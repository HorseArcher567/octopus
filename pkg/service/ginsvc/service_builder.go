package ginsvc

import (
	"github.com/gin-gonic/gin"
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
	Gin struct {
		Enabled    bool              `json:"enabled,omitempty" yaml:"enabled,omitempty"`
		Mode       string            `json:"mode,omitempty" yaml:"mode,omitempty"`
		Name       string            `json:"name,omitempty" yaml:"name,omitempty"`
		Address    string            `json:"address,omitempty" yaml:"address,omitempty"`
		Prometheus config.Prometheus `json:"prometheus,omitempty" yaml:"prometheus,omitempty"`
	} `json:"gin,omitempty" yaml:"gin,omitempty"`
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

		gin.SetMode(conf.Gin.Mode)
		engine := gin.New()
		svc := &Service{
			enabled: conf.Gin.Enabled,
			name:    conf.Gin.Name,
			server: &http.Server{
				Handler: engine,
			},
			address: conf.Gin.Address,
		}
		if conf.Gin.Prometheus.Server.Enabled {
			svc.metrics = metrics.NewGinServerMetrics(conf.Gin.Prometheus.Server.Namespace,
				conf.Gin.Prometheus.Server.Subsystem)
			promsvc.MustRegister(svc.metrics)

			if conf.Gin.Prometheus.Server.CountsHandlingTime {
				svc.metrics.EnableCountsHandlingTime()
			}
			engine.Use(svc.metrics.MiddlewareHandler())
		}

		b.service = svc
	})

	return b.service
}
