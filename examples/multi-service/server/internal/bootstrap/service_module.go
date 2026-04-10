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

	if err := b.Container().Resolve(&userRepo); err != nil {
		return err
	}
	if err := b.Container().Resolve(&orderRepo); err != nil {
		return err
	}
	if err := b.Container().Resolve(&productRepo); err != nil {
		return err
	}

	if err := b.Container().Provide(service.NewUserService(userRepo)); err != nil {
		return err
	}
	if err := b.Container().Provide(service.NewOrderService(orderRepo)); err != nil {
		return err
	}
	return b.Container().Provide(service.NewProductService(productRepo))
}
