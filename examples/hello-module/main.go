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

func (m *helloModule) RegisterHTTP(_ context.Context, r app.HTTPRegistrar) error {
	return r.RegisterHTTP(func(engine *api.Engine) {
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
