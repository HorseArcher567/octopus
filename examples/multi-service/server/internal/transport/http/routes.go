package http

import (
	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(engine *api.Engine, user *UserHandler, order *OrderHandler, product *ProductHandler) {
	engine.GET("/hello", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "hello from apiServer"})
	})

	engine.GET("/users/:id", user.GetUser)
	engine.POST("/users", user.CreateUser)

	engine.GET("/orders/:id", order.GetOrder)
	engine.POST("/orders", order.CreateOrder)

	engine.GET("/products/:id", product.GetProduct)
	engine.GET("/products", product.ListProducts)
}
