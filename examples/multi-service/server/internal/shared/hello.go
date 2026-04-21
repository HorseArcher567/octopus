package shared

import (
	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/HorseArcher567/octopus/pkg/assemble"
	"github.com/gin-gonic/gin"
)

func AssembleHello(ctx *assemble.Context) error {
	return ctx.RegisterAPI(func(engine *api.Engine) {
		engine.GET("/hello", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "hello from apiServer"})
		})
	})
}
