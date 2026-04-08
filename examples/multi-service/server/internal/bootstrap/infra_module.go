package bootstrap

import (
	"context"

	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/repository"
	"github.com/HorseArcher567/octopus/pkg/app"
)

const mysqlPrimary = "primary"

type InfraModule struct {
}

func NewInfraModule() *InfraModule {
	return &InfraModule{}
}

func (m *InfraModule) ID() string { return "infra" }

func (m *InfraModule) Build(_ context.Context, b app.BuildContext) error {
	db, err := b.MySQL(mysqlPrimary)
	if err != nil {
		return err
	}

	if err := b.Provide(repository.NewUserRepository(db)); err != nil {
		return err
	}
	if err := b.Provide(repository.NewOrderRepository(db)); err != nil {
		return err
	}
	return b.Provide(repository.NewProductRepository(db))
}
