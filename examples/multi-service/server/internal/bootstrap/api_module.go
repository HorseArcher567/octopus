package bootstrap

import (
	"context"

	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/service"
	httpx "github.com/HorseArcher567/octopus/examples/multi-service/server/internal/transport/http"
	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/HorseArcher567/octopus/pkg/app"
)

type APIModule struct{}

func NewAPIModule() *APIModule {
	return &APIModule{}
}

func (m *APIModule) ID() string { return "api" }

func (m *APIModule) DependsOn() []string { return []string{"service"} }

func (m *APIModule) RegisterHTTP(_ context.Context, r app.HTTPRegistrar) error {
	var userSvc *service.UserService
	var orderSvc *service.OrderService
	var productSvc *service.ProductService

	r.MustResolve(&userSvc)
	r.MustResolve(&orderSvc)
	r.MustResolve(&productSvc)

	log := r.Logger()
	return r.RegisterHTTP(func(engine *api.Engine) {
		httpx.RegisterRoutes(
			engine,
			httpx.NewUserHandler(userSvc, log),
			httpx.NewOrderHandler(orderSvc, log),
			httpx.NewProductHandler(productSvc, log),
		)
	})
}
