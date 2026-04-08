package app

import (
	"fmt"
	"time"

	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/HorseArcher567/octopus/pkg/config"
	"github.com/HorseArcher567/octopus/pkg/etcd"
	"github.com/HorseArcher567/octopus/pkg/job"
	"github.com/HorseArcher567/octopus/pkg/resource"
	"github.com/HorseArcher567/octopus/pkg/rpc"
	"github.com/HorseArcher567/octopus/pkg/xlog"
)

type frameworkConfig struct {
	Logger           xlog.Config        `yaml:"logger" json:"logger" toml:"logger"`
	Etcd             *etcd.Config       `yaml:"etcd" json:"etcd" toml:"etcd"`
	RPCServer        *rpc.ServerConfig  `yaml:"rpcServer" json:"rpcServer" toml:"rpcServer"`
	RPCClientOptions *rpc.ClientOptions `yaml:"rpcClientOptions" json:"rpcClientOptions" toml:"rpcClientOptions"`
	APIServer        *api.ServerConfig  `yaml:"apiServer" json:"apiServer" toml:"apiServer"`
	Resources        *resource.Config   `yaml:"resources" json:"resources" toml:"resources"`
}

// FromConfig builds an App from framework config with strict field decoding.
func FromConfig(cfg *config.Config, opts ...Option) (*App, error) {
	if cfg == nil {
		return nil, fmt.Errorf("app: config cannot be nil")
	}

	fc, err := decodeFrameworkConfig(cfg)
	if err != nil {
		return nil, err
	}

	log, err := xlog.New(&fc.Logger)
	if err != nil {
		return nil, fmt.Errorf("app: logger: %w", err)
	}

	baseOpts := make([]Option, 0, len(opts)+4)
	baseOpts = append(baseOpts, WithJobRuntime(job.NewScheduler(log)))
	var resourceRuntime ResourceRuntime

	rpcRuntime, err := rpc.NewRuntime(log, &rpc.RuntimeConfig{
		Server:        fc.RPCServer,
		ClientOptions: fc.RPCClientOptions,
		Etcd:          fc.Etcd,
	})
	if err != nil {
		_ = log.Close()
		return nil, fmt.Errorf("app: rpc runtime: %w", err)
	}
	baseOpts = append(baseOpts, WithRPCRuntime(rpcRuntime))

	if fc.APIServer != nil {
		httpRuntime, err := api.NewServer(log, fc.APIServer)
		if err != nil {
			_ = rpcRuntime.Close()
			_ = log.Close()
			return nil, fmt.Errorf("app: http runtime: %w", err)
		}
		baseOpts = append(baseOpts, WithHTTPRuntime(httpRuntime))
	}

	if fc.Resources != nil {
		resourceRuntime, err = resource.New(fc.Resources)
		if err != nil {
			_ = rpcRuntime.Close()
			_ = log.Close()
			return nil, fmt.Errorf("app: resource runtime: %w", err)
		}
		baseOpts = append(baseOpts, WithResourceRuntime(resourceRuntime))
	}

	if rawTimeout, ok := cfg.Get("shutdownTimeout"); ok {
		timeout, err := parseShutdownTimeout(rawTimeout)
		if err != nil {
			_ = rpcRuntime.Close()
			if resourceRuntime != nil {
				_ = resourceRuntime.Close()
			}
			_ = log.Close()
			return nil, err
		}
		baseOpts = append(baseOpts, WithShutdownTimeout(timeout))
	}

	baseOpts = append(baseOpts, opts...)
	return New(log, baseOpts...), nil
}

// Load reads configPath and builds an App from framework config.
func Load(configPath string, opts ...Option) (*App, error) {
	if configPath == "" {
		return nil, fmt.Errorf("app: config path is required")
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	return FromConfig(cfg, opts...)
}

func parseShutdownTimeout(value any) (time.Duration, error) {
	switch v := value.(type) {
	case time.Duration:
		return v, nil
	case string:
		d, err := time.ParseDuration(v)
		if err != nil {
			return 0, fmt.Errorf("app: invalid shutdownTimeout %q: %w", v, err)
		}
		return d, nil
	case int:
		return time.Duration(v) * time.Second, nil
	case int64:
		return time.Duration(v) * time.Second, nil
	case float64:
		return time.Duration(v * float64(time.Second)), nil
	default:
		return 0, fmt.Errorf("app: invalid shutdownTimeout type %T", value)
	}
}

func decodeFrameworkConfig(cfg *config.Config) (*frameworkConfig, error) {
	fc := &frameworkConfig{}

	decodeKey := func(key string, target any) error {
		if !cfg.Has(key) {
			return nil
		}
		if err := cfg.UnmarshalKeyStrict(key, target); err != nil {
			return fmt.Errorf("app: invalid %s config: %w", key, err)
		}
		return nil
	}

	if err := decodeKey("logger", &fc.Logger); err != nil {
		return nil, err
	}
	if err := decodeKey("etcd", &fc.Etcd); err != nil {
		return nil, err
	}
	if err := decodeKey("rpcServer", &fc.RPCServer); err != nil {
		return nil, err
	}
	if err := decodeKey("rpcClientOptions", &fc.RPCClientOptions); err != nil {
		return nil, err
	}
	if err := decodeKey("apiServer", &fc.APIServer); err != nil {
		return nil, err
	}
	if err := decodeKey("resources", &fc.Resources); err != nil {
		return nil, err
	}

	return fc, nil
}
