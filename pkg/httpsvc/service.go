package httpsvc

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/k8s-practice/octopus/pkg/log"
	"github.com/k8s-practice/octopus/pkg/service"
	"github.com/k8s-practice/octopus/pkg/util/structure"
	"net"
	"net/http"
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
		log.Panicf("%s service is not a http service", name)
		return nil
	}
	return svc
}

type Config struct {
	Http []struct {
		Enabled bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
		Name    string `json:"name,omitempty" yaml:"name,omitempty"`
		Address string `json:"address,omitempty" yaml:"address,omitempty"`
	} `json:"http,omitempty" yaml:"http,omitempty"`
}

type Service struct {
	name    string
	server  *http.Server
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
	for i := range config.Http {
		if !config.Http[i].Enabled {
			continue
		}

		gin.SetMode(gin.ReleaseMode)
		svc := &Service{
			name: config.Http[i].Name,
			server: &http.Server{
				Handler: gin.New(),
			},
			address: config.Http[i].Address,
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

func (svc *Service) Stop(ctx context.Context) {
	log.Errorln(svc.server.Shutdown(ctx))
}

func (svc *Service) Router() *gin.Engine {
	return svc.server.Handler.(*gin.Engine)
}
