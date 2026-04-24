package assemble

import (
	"fmt"

	"github.com/HorseArcher567/octopus/pkg/api"
)

func setupAPI(c *setupContext) error {
	if _, ok := c.get("apiServer"); !ok {
		return nil
	}
	var cfg api.ServerConfig
	if err := c.decodeStruct("apiServer", &cfg); err != nil {
		return err
	}
	log, err := selectLogger(cfg.Logger, c.state.log, c.state.store)
	if err != nil {
		return fmt.Errorf("assemble: apiServer.logger: %w", err)
	}
	server, err := api.NewServer(log, &cfg)
	if err != nil {
		return fmt.Errorf("assemble: api server: %w", err)
	}
	c.state.api = server
	return nil
}
