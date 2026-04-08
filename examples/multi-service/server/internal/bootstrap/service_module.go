package bootstrap

import (
	"context"

	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/repository"
	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/service"
	"github.com/HorseArcher567/octopus/pkg/app"
)

type ServiceModule struct{}

func NewServiceModule() *ServiceModule {
	return &ServiceModule{}
}

func (m *ServiceModule) ID() string { return "service" }

func (m *ServiceModule) DependsOn() []string { return []string{"infra"} }

func (m *ServiceModule) Build(_ context.Context, b app.BuildContext) error {
	var userRepo repository.UserRepository
	var orderRepo repository.OrderRepository
	var productRepo repository.ProductRepository

	b.MustResolve(&userRepo)
	b.MustResolve(&orderRepo)
	b.MustResolve(&productRepo)

	if err := b.Provide(service.NewUserService(userRepo)); err != nil {
		return err
	}
	if err := b.Provide(service.NewOrderService(orderRepo)); err != nil {
		return err
	}
	return b.Provide(service.NewProductService(productRepo))
}
