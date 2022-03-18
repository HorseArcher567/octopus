package ginsvc

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
	Gin struct {
		Enabled bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
		Mode    string `json:"mode,omitempty" yaml:"mode,omitempty"`
		Name    string `json:"name,omitempty" yaml:"name,omitempty"`
		Address string `json:"address,omitempty" yaml:"address,omitempty"`
	} `json:"gin,omitempty" yaml:"gin,omitempty"`
}

func (builder *Builder) Build(bootConfig map[interface{}]interface{}, tag string) service.Entry {
	var conf Config
	if err := structure.UnmarshalWithTag(bootConfig, &conf, tag); err != nil {
		log.Panicln(err)
		return nil
	}
	if !conf.Gin.Enabled {
		return nil
	}

	gin.SetMode(conf.Gin.Mode)
	singleton = &Service{
		name: conf.Gin.Name,
		server: &http.Server{
			Handler: gin.New(),
		},
		address: conf.Gin.Address,
	}

	return singleton
}
