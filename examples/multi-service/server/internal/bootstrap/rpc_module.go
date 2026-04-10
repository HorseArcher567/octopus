package bootstrap

import (
	"context"

	"github.com/HorseArcher567/octopus/examples/multi-service/proto/pb"
	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/service"
	grpcx "github.com/HorseArcher567/octopus/examples/multi-service/server/internal/transport/grpc"
	"github.com/HorseArcher567/octopus/pkg/app"
	"google.golang.org/grpc"
)

type RPCModule struct{}

func NewRPCModule() *RPCModule {
	return &RPCModule{}
}

func (m *RPCModule) ID() string { return "rpc" }

func (m *RPCModule) DependsOn() []string { return []string{"service"} }

func (m *RPCModule) RegisterRPC(_ context.Context, r app.RPCRegistrar) error {
	var userSvc *service.UserService
	var orderSvc *service.OrderService
	var productSvc *service.ProductService

	if err := r.Resolve(&userSvc); err != nil {
		return err
	}
	if err := r.Resolve(&orderSvc); err != nil {
		return err
	}
	if err := r.Resolve(&productSvc); err != nil {
		return err
	}

	log := r.Logger()
	return r.RegisterRPC(func(s *grpc.Server) {
		pb.RegisterUserServer(s, grpcx.NewUserHandler(userSvc, log))
		pb.RegisterOrderServer(s, grpcx.NewOrderHandler(orderSvc, log))
		pb.RegisterProductServer(s, grpcx.NewProductHandler(productSvc, log))
	})
}
