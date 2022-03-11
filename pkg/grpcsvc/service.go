package grpcsvc

import (
	"context"
	"github.com/k8s-practice/octopus/pkg/log"
	"github.com/k8s-practice/octopus/pkg/service"
	"github.com/k8s-practice/octopus/pkg/util/structure"
	"google.golang.org/grpc"
	"net"
)

func init() {
	service.RegisterBuilder(&Builder{})
}

func MustGetService(name string) *Service {
	entry := service.Get(name)
	if entry == nil {
		log.Panicf("%s entry not exist", name)
		return nil
	}

	svc, ok := entry.(*Service)
	if !ok {
		log.Panicf("%s service is not a grpc service", name)
		return nil
	}
	return svc
}

type Config struct {
	Grpc []struct {
		Enabled bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
		Name    string `json:"name,omitempty" yaml:"name,omitempty"`
		Address string `json:"address,omitempty" yaml:"address,omitempty"`
	} `json:"simple,omitempty" yaml:"simple,omitempty"`
}

type Service struct {
	name    string
	server  *grpc.Server
	address string
}

type Builder struct {
}

func (builder *Builder) Build(bootConfig map[interface{}]interface{}) []service.Entry {
	var config Config
	if err := structure.UnmarshalWithTag(bootConfig, &config, "yaml"); err != nil {
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

func (svc *Service) Name() string {
	return svc.name
}

func (svc *Service) Run(_ context.Context) {
	listener, err := net.Listen("tcp", svc.address)
	if err != nil {
		log.Panicln(err)
		return
	}
	log.Errorln(svc.server.Serve(listener))
}

func (svc *Service) Stop(_ context.Context) {
	svc.server.GracefulStop()
}

type RegisterFunc func(*grpc.Server)

func (svc *Service) Register(f RegisterFunc) {
	f(svc.server)
}
