package http

import (
	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(engine *api.Engine) {
	engine.GET("/hello", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "hello from apiServer"})
	})
}
