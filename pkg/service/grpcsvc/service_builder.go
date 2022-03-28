package grpcsvc

import (
	"github.com/k8s-practice/octopus/pkg/log"
	"github.com/k8s-practice/octopus/pkg/prometheus/metrics"
	"github.com/k8s-practice/octopus/pkg/service"
	"github.com/k8s-practice/octopus/pkg/service/config"
	"github.com/k8s-practice/octopus/pkg/service/promsvc"
	"github.com/k8s-practice/octopus/pkg/util/structure"
	"google.golang.org/grpc"
)

type Config struct {
	Grpc struct {
		Enabled    bool              `json:"enabled,omitempty" yaml:"enabled,omitempty"`
		Name       string            `json:"name,omitempty" yaml:"name,omitempty"`
		Address    string            `json:"address,omitempty" yaml:"address,omitempty"`
		Prometheus config.Prometheus `json:"prometheus,omitempty" yaml:"prometheus,omitempty"`
	} `json:"grpc,omitempty" yaml:"grpc,omitempty"`
}

type Builder struct {
}

func (builder *Builder) Build(bootConfig map[interface{}]interface{}, tag string) service.Entry {
	var conf Config
	if err := structure.UnmarshalWithTag(bootConfig, &conf, tag); err != nil {
		log.Panicln(err)
		return nil
	}

	singleton = &Service{
		enabled: conf.Grpc.Enabled,
		name:    conf.Grpc.Name,
		address: conf.Grpc.Address,
	}
	if conf.Grpc.Prometheus.Server.Enabled {
		singleton.metrics = metrics.NewGrpcServerMetrics(conf.Grpc.Prometheus.Server.Namespace,
			conf.Grpc.Prometheus.Server.Subsystem)
		promsvc.MustRegister(singleton.metrics)

		singleton.server = grpc.NewServer(grpc.StreamInterceptor(singleton.metrics.StreamServerInterceptor()),
			grpc.UnaryInterceptor(singleton.metrics.UnaryServerInterceptor()))
		singleton.beforeServe = append(singleton.beforeServe, func() {
			if conf.Grpc.Prometheus.Server.CountsHandlingTime {
				singleton.metrics.EnableCountsHandlingTime()
			}
			singleton.metrics.InitializeMetrics(singleton.server)
		})
	} else {
		singleton.server = grpc.NewServer()
	}

	return singleton
}