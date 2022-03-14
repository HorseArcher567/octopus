package grpcsvc

import (
	"github.com/k8s-practice/octopus/pkg/log"
	"github.com/k8s-practice/octopus/pkg/service"
	"github.com/k8s-practice/octopus/pkg/util/structure"
	"google.golang.org/grpc"
)

type Config struct {
	Grpc []struct {
		Enabled bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
		Name    string `json:"name,omitempty" yaml:"name,omitempty"`
		Address string `json:"address,omitempty" yaml:"address,omitempty"`
	} `json:"simple,omitempty" yaml:"simple,omitempty"`
}

type Builder struct {
}

func (builder *Builder) Build(bootConfig map[interface{}]interface{}, tag string) []service.Entry {
	var config Config
	if err := structure.UnmarshalWithTag(bootConfig, &config, tag); err != nil {
		log.Panicln(err)
		return nil
	}

	var services []service.Entry
	for i := range config.Grpc {
		if !config.Grpc[i].Enabled {
			continue
		}

		svc := &Service{
			name:    config.Grpc[i].Name,
			server:  grpc.NewServer(),
			address: config.Grpc[i].Address,
		}
		services = append(services, svc)
	}

	return services
}
