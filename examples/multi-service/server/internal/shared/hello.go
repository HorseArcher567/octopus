package shared

import (
	"fmt"
	"strings"

	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/HorseArcher567/octopus/pkg/assemble"
	"github.com/HorseArcher567/octopus/pkg/store"
	"github.com/gin-gonic/gin"
)

type HelloConfig struct {
	Message string `yaml:"message"`
}

func SetupHello() assemble.SetupStep {
	return assemble.SetupStep{
		Name: "hello",
		Run: func(ctx *assemble.SetupContext) error {
			cfg, err := assemble.DecodeSetupConfig[HelloConfig](ctx, "hello")
			if err != nil {
				return err
			}
			message := strings.TrimSpace(cfg.Message)
			if message == "" {
				return fmt.Errorf("hello.message is required")
			}
			return ctx.Provide("default", message)
		},
	}
}

func AssembleHello(ctx *assemble.Context) error {
	message, err := store.GetNamed[string](ctx.Store(), "default")
	if err != nil {
		return err
	}
	return ctx.RegisterAPI(func(engine *api.Engine) {
		engine.GET("/hello", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": message})
		})
	})
}
