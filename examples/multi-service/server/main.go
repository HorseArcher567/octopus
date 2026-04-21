package main

import (
	"context"
	"flag"

	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/order"
	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/product"
	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/shared"
	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/user"
	"github.com/HorseArcher567/octopus/pkg/assemble"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	configFile := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	a, err := assemble.Load(
		*configFile,
		assemble.With(
			shared.AssembleHello,
			user.Assemble,
			order.Assemble,
			product.Assemble,
		),
	)
	if err != nil {
		panic(err)
	}
	if err := a.Run(context.Background()); err != nil {
		panic(err)
	}
}
