package bootstrap

import (
	"context"
	"errors"

	"github.com/HorseArcher567/octopus/examples/multi-service/proto/pb"
	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/service"
	grpcx "github.com/HorseArcher567/octopus/examples/multi-service/server/internal/transport/grpc"
	"github.com/HorseArcher567/octopus/pkg/app"
	"google.golang.org/grpc"
)

type RPCModule struct {
	infra *InfraModule
}

func NewRPCModule(infra *InfraModule) *RPCModule {
	return &RPCModule{infra: infra}
}

func (m *RPCModule) ID() string { return "rpc" }

func (m *RPCModule) DependsOn() []string { return []string{"infra"} }

func (m *RPCModule) Init(_ context.Context, rt app.Runtime) error {
	if m.infra == nil {
		return errors.New("infra module is required")
	}

	userSvc := service.NewUserService(m.infra.UserRepo())
	orderSvc := service.NewOrderService(m.infra.OrderRepo())
	productSvc := service.NewProductService(m.infra.ProductRepo())

	log := rt.Logger()
	rt.RegisterRPC(func(s *grpc.Server) {
		pb.RegisterUserServer(s, grpcx.NewUserHandler(userSvc, log))
		pb.RegisterOrderServer(s, grpcx.NewOrderHandler(orderSvc, log))
		pb.RegisterProductServer(s, grpcx.NewProductHandler(productSvc, log))
	})
	return nil
}

func (m *RPCModule) Close(_ context.Context) error {
	return nil
}
