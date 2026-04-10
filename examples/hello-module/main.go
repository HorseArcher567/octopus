package main

import (
	"context"
	"flag"

	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/HorseArcher567/octopus/pkg/app"
	"github.com/gin-gonic/gin"
)

type helloModule struct{}

func (m *helloModule) ID() string { return "hello" }

func (m *helloModule) RegisterAPI(_ context.Context, r app.APIRegistrar) error {
	return r.RegisterAPI(func(engine *api.Engine) {
		engine.GET("/hello", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "hello from hello-module"})
		})
	})
}

func main() {
	configFile := flag.String("config", "config.yaml", "config file path")
	flag.Parse()

	app.MustRun(*configFile, []app.Module{&helloModule{}})
}
