package assemble

import (
	"fmt"

	"github.com/HorseArcher567/octopus/pkg/config"
	"github.com/HorseArcher567/octopus/pkg/store"
	"github.com/HorseArcher567/octopus/pkg/xlog"
)

type state struct {
	log   *xlog.Logger
	store store.Store

	api apiServer
	rpc rpcServer
	job jobScheduler
}

// setupContext is the internal setup-time context used by builtin setup steps.
type setupContext struct {
	cfg   *config.Config
	state *state
}

// builtinSetupStep is an internal framework-controlled setup step.
type builtinSetupStep struct {
	name string
	run  func(*setupContext) error
}

var builtinSetupSteps = []builtinSetupStep{
	{name: "loggers", run: setupLoggers},
	{name: "app-logger", run: selectAppLogger},
	{name: "etcd", run: setupEtcd},
	{name: "mysql", run: setupMySQL},
	{name: "redis", run: setupRedis},
	{name: "rpc-resolver", run: setupRPCResolver},
	{name: "api", run: setupAPI},
	{name: "rpc", run: setupRPC},
	{name: "jobs", run: setupJobs},
}

func setup(cfg *config.Config) (*state, error) {
	if cfg == nil {
		return nil, fmt.Errorf("assemble: config cannot be nil")
	}

	st := &state{store: store.New()}
	ctx := &setupContext{cfg: cfg, state: st}
	if err := runBuiltinSetupSteps(ctx, builtinSetupSteps); err != nil {
		_ = st.store.Close()
		return nil, err
	}
	return st, nil
}

func runBuiltinSetupSteps(ctx *setupContext, steps []builtinSetupStep) error {
	for _, step := range steps {
		if err := step.run(ctx); err != nil {
			return fmt.Errorf("assemble: setup %s: %w", step.name, err)
		}
	}
	return nil
}

func (c *setupContext) decodeStruct(key string, out any) error {
	if err := c.cfg.UnmarshalKey(key, out); err != nil {
		return fmt.Errorf("decode config %q: %w", key, err)
	}
	return nil
}

func (c *setupContext) get(key string) (any, bool) {
	return c.cfg.Get(key)
}

func (c *setupContext) provide(name string, value any, opts ...store.SetOption) error {
	if err := c.state.store.SetNamed(name, value, opts...); err != nil {
		return fmt.Errorf("provide %q: %w", name, err)
	}
	return nil
}
