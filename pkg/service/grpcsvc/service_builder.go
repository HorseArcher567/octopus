package grpcsvc

import (
	"github.com/k8s-practice/octopus/pkg/log"
	"github.com/k8s-practice/octopus/pkg/prometheus/metrics"
	"github.com/k8s-practice/octopus/pkg/service"
	"github.com/k8s-practice/octopus/pkg/service/config"
	"github.com/k8s-practice/octopus/pkg/service/promsvc"
	"github.com/k8s-practice/octopus/pkg/util/structure"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"sync"
)

type Config struct {
	Grpc struct {
		Enabled       bool                 `json:"enabled,omitempty" yaml:"enabled,omitempty"`
		Name          string               `json:"name,omitempty" yaml:"name,omitempty"`
		Address       string               `json:"address,omitempty" yaml:"address,omitempty"`
		Prometheus    config.Prometheus    `json:"prometheus,omitempty" yaml:"prometheus,omitempty"`
		OpenTelemetry config.OpenTelemetry `json:"openTelemetry,omitempty" yaml:"openTelemetry,omitempty"`
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

		var usi []grpc.UnaryServerInterceptor
		var ssi []grpc.StreamServerInterceptor
		svc := &Service{
			enabled: conf.Grpc.Enabled,
			name:    conf.Grpc.Name,
			address: conf.Grpc.Address,
		}

		if conf.Grpc.OpenTelemetry.Enabled {
			usi = append(usi, otelgrpc.UnaryServerInterceptor())
			ssi = append(ssi, otelgrpc.StreamServerInterceptor())
		}

		if conf.Grpc.Prometheus.Server.Enabled {
			svc.metrics = metrics.NewGrpcServerMetrics(conf.Grpc.Prometheus.Server.Namespace,
				conf.Grpc.Prometheus.Server.Subsystem)
			promsvc.MustRegister(svc.metrics)

			usi = append(usi, svc.metrics.UnaryServerInterceptor())
			ssi = append(ssi, svc.metrics.StreamServerInterceptor())
			svc.beforeServe = append(svc.beforeServe, func() {
				if conf.Grpc.Prometheus.Server.CountsHandlingTime {
					svc.metrics.EnableCountsHandlingTime()
				}
				svc.metrics.InitializeMetrics(svc.server)
			})
		}
		svc.server = grpc.NewServer(grpc.ChainUnaryInterceptor(usi...),
			grpc.ChainStreamInterceptor(ssi...))

		b.service = svc
	})

	return b.service
}
