package grpcsvc

import (
	"github.com/k8s-practice/octopus/pkg/log"
	"github.com/k8s-practice/octopus/pkg/prometheus/metrics"
	"github.com/k8s-practice/octopus/pkg/service"
	"github.com/k8s-practice/octopus/pkg/service/config"
	"github.com/k8s-practice/octopus/pkg/service/promsvc"
	"github.com/k8s-practice/octopus/pkg/util/structure"
	"google.golang.org/grpc"
	"sync"
)

type Config struct {
	Grpc struct {
		Enabled    bool              `json:"enabled,omitempty" yaml:"enabled,omitempty"`
		Name       string            `json:"name,omitempty" yaml:"name,omitempty"`
		Address    string            `json:"address,omitempty" yaml:"address,omitempty"`
		Prometheus config.Prometheus `json:"prometheus,omitempty" yaml:"prometheus,omitempty"`
	} `json:"grpc,omitempty" yaml:"grpc,omitempty"`
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

		svc := &Service{
			enabled: conf.Grpc.Enabled,
			name:    conf.Grpc.Name,
			address: conf.Grpc.Address,
		}
		if conf.Grpc.Prometheus.Server.Enabled {
			svc.metrics = metrics.NewGrpcServerMetrics(conf.Grpc.Prometheus.Server.Namespace,
				conf.Grpc.Prometheus.Server.Subsystem)
			promsvc.MustRegister(svc.metrics)

			svc.server = grpc.NewServer(grpc.StreamInterceptor(svc.metrics.StreamServerInterceptor()),
				grpc.UnaryInterceptor(svc.metrics.UnaryServerInterceptor()))
			svc.beforeServe = append(svc.beforeServe, func() {
				if conf.Grpc.Prometheus.Server.CountsHandlingTime {
					svc.metrics.EnableCountsHandlingTime()
				}
				svc.metrics.InitializeMetrics(svc.server)
			})
		} else {
			svc.server = grpc.NewServer()
		}

		b.service = svc
	})

	return b.service
}
