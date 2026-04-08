package main

import (
	"flag"

	"github.com/HorseArcher567/octopus/examples/multi-service/server/internal/bootstrap"
	"github.com/HorseArcher567/octopus/pkg/app"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	configFile := flag.String("config", "configs/config.dev.yaml", "配置文件路径")
	flag.Parse()

	app.MustRun(*configFile, []app.Module{
		bootstrap.NewInfraModule(),
		bootstrap.NewServiceModule(),
		bootstrap.NewRPCModule(),
		bootstrap.NewAPIModule(),
	})
}
