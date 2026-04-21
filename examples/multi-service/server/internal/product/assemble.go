package product

import (
	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/shared"
	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/HorseArcher567/octopus/pkg/assemble"
	"google.golang.org/grpc"
)

func Assemble(ctx *assemble.Context) error {
	db, err := shared.PrimaryDB(ctx)
	if err != nil { return err }
	repo := NewRepository(db)
	svc := NewService(repo)
	log := ctx.Logger()
	if err := ctx.RegisterRPC(func(s *grpc.Server) { RegisterGRPC(s, NewGRPCHandler(svc, log)) }); err != nil { return err }
	return ctx.RegisterAPI(func(engine *api.Engine) { RegisterHTTP(engine, NewHTTPHandler(svc, log)) })
}
