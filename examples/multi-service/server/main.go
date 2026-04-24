package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/order"
	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/product"
	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/shared"
	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/user"
	"github.com/HorseArcher567/octopus/pkg/assemble"
)

func main() {
	configFile := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	a, err := assemble.Load(
		*configFile,
		assemble.WithStartupHooks(shared.InitSchema),
		assemble.WithDomains(
			user.Register,
			order.Register,
			product.Register,
		),
	)
	if err != nil {
		panic(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := a.Run(ctx); err != nil {
		panic(err)
	}
}
