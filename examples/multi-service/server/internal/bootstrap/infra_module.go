package bootstrap

import (
	"context"

	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/repository"
	"github.com/HorseArcher567/octopus/pkg/app"
)

const mysqlPrimary = "primary"

type InfraModule struct {
	userRepo    repository.UserRepository
	orderRepo   repository.OrderRepository
	productRepo repository.ProductRepository
}

func NewInfraModule() *InfraModule {
	return &InfraModule{}
}

func (m *InfraModule) ID() string { return "infra" }

func (m *InfraModule) Init(_ context.Context, rt app.Runtime) error {
	db, err := rt.MySQL(mysqlPrimary)
	if err != nil {
		return err
	}

	m.userRepo = repository.NewUserRepository(db)
	m.orderRepo = repository.NewOrderRepository(db)
	m.productRepo = repository.NewProductRepository(db)
	return nil
}

func (m *InfraModule) Close(_ context.Context) error {
	return nil
}

func (m *InfraModule) UserRepo() repository.UserRepository {
	return m.userRepo
}

func (m *InfraModule) OrderRepo() repository.OrderRepository {
	return m.orderRepo
}

func (m *InfraModule) ProductRepo() repository.ProductRepository {
	return m.productRepo
}
