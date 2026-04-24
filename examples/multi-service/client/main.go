package main

import (
	"context"
	"flag"

	"github.com/HorseArcher567/octopus/examples/multi-service/client/internal/jobs"
	"github.com/HorseArcher567/octopus/pkg/assemble"
)

func main() {
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	target := flag.String("target", "etcd:///multi-service-demo", "RPC target")
	apiURL := flag.String("api", "http://127.0.0.1:8090/hello", "API URL")
	flag.Parse()

	a, err := assemble.Load(
		*configPath,
		assemble.WithDomains(jobs.Register(*target, *apiURL)),
	)
	if err != nil {
		panic(err)
	}

	if err := a.Run(context.Background()); err != nil {
		panic(err)
	}
}
