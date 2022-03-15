package httpsvc

import (
	"github.com/gin-gonic/gin"
	"github.com/k8s-practice/octopus/pkg/log"
	"github.com/k8s-practice/octopus/pkg/service"
	"github.com/k8s-practice/octopus/pkg/util/structure"
	"net/http"
)

type Builder struct {
}

type Config struct {
	Http []struct {
		Enabled bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
		Name    string `json:"name,omitempty" yaml:"name,omitempty"`
		Address string `json:"address,omitempty" yaml:"address,omitempty"`
	} `json:"http,omitempty" yaml:"http,omitempty"`
}

func (builder *Builder) Build(bootConfig map[interface{}]interface{}, tag string) []service.Entry {
	var config Config
	if err := structure.UnmarshalWithTag(bootConfig, &config, tag); err != nil {
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
