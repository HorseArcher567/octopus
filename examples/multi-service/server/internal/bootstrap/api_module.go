package bootstrap

import (
	"context"

	httpx "github.com/HorseArcher567/octopus/examples/multi-service/server/internal/transport/http"
	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/HorseArcher567/octopus/pkg/app"
)

type APIModule struct{}

func NewAPIModule() *APIModule {
	return &APIModule{}
}

func (m *APIModule) ID() string { return "api" }

func (m *APIModule) Init(_ context.Context, rt app.Runtime) error {
	rt.RegisterHTTP(func(engine *api.Engine) {
		httpx.RegisterRoutes(engine)
	})
	return nil
}

func (m *APIModule) Close(_ context.Context) error {
	return nil
}
