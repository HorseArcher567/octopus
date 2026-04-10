package app

// This file contains the bootstrap and assembly path that constructs framework
// runtimes from configuration before creating an App.

import (
	"fmt"
	"time"

	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/HorseArcher567/octopus/pkg/config"
	"github.com/HorseArcher567/octopus/pkg/etcd"
	"github.com/HorseArcher567/octopus/pkg/health"
	"github.com/HorseArcher567/octopus/pkg/job"
	"github.com/HorseArcher567/octopus/pkg/resource"
	"github.com/HorseArcher567/octopus/pkg/rpc"
	"github.com/HorseArcher567/octopus/pkg/telemetry"
	obsmetrics "github.com/HorseArcher567/octopus/pkg/telemetry/metrics"
	obstrace "github.com/HorseArcher567/octopus/pkg/telemetry/trace"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	"github.com/gin-gonic/gin"
)

// Bootstrap contains fully assembled runtime dependencies used to construct an App.
type Bootstrap struct {
	Logger          *xlog.Logger
	API             APIRuntime
	RPC             RPCRuntime
	Jobs            JobRuntime
	Resources       ResourceRuntime
	Telemetry       *telemetry.Runtime
	Health          *health.Runtime
	ShutdownTimeout time.Duration
}

// BootstrapOption customizes how framework runtimes are assembled from config.
type BootstrapOption func(*BootstrapConfig)

// BootstrapConfig collects child-runtime options during assembly.
type BootstrapConfig struct {
	APIOptions       []api.Option
	RPCServerOptions []rpc.Option
	ResourceOptions  []resource.Option
	TelemetryOptions []telemetry.Option
}

// WithAPIOptions passes options to api.NewServer during assembly.
func WithAPIOptions(opts ...api.Option) BootstrapOption {
	return func(c *BootstrapConfig) {
		c.APIOptions = append(c.APIOptions, opts...)
	}
}

// WithRPCServerOptions passes options to rpc.NewRuntime/rpc.NewServer during assembly.
func WithRPCServerOptions(opts ...rpc.Option) BootstrapOption {
	return func(c *BootstrapConfig) {
		c.RPCServerOptions = append(c.RPCServerOptions, opts...)
	}
}

// WithResourceOptions passes options to resource.New during assembly.
func WithResourceOptions(opts ...resource.Option) BootstrapOption {
	return func(c *BootstrapConfig) {
		c.ResourceOptions = append(c.ResourceOptions, opts...)
	}
}

// WithTelemetryOptions passes options to telemetry runtime assembly.
func WithTelemetryOptions(opts ...telemetry.Option) BootstrapOption {
	return func(c *BootstrapConfig) {
		c.TelemetryOptions = append(c.TelemetryOptions, opts...)
	}
}

// rollbackStack records closers for partial assembly rollback.
type rollbackStack struct {
	closers []func() error
}

// Add registers a closer to be executed during rollback.
func (r *rollbackStack) Add(fn func() error) {
	if fn != nil {
		r.closers = append(r.closers, fn)
	}
}

// Rollback closes registered resources in reverse registration order.
func (r *rollbackStack) Rollback() {
	for i := len(r.closers) - 1; i >= 0; i-- {
		_ = r.closers[i]()
	}
}

// frameworkConfig contains the framework-owned configuration sections decoded
// from the root config.
type frameworkConfig struct {
	Logger           xlog.Config        `yaml:"logger" json:"logger" toml:"logger"`
	Etcd             *etcd.Config       `yaml:"etcd" json:"etcd" toml:"etcd"`
	RPCServer        *rpc.ServerConfig  `yaml:"rpcServer" json:"rpcServer" toml:"rpcServer"`
	RPCClientOptions *rpc.ClientOptions `yaml:"rpcClientOptions" json:"rpcClientOptions" toml:"rpcClientOptions"`
	APIServer        *api.ServerConfig  `yaml:"apiServer" json:"apiServer" toml:"apiServer"`
	Resources        *resource.Config   `yaml:"resources" json:"resources" toml:"resources"`
	Health           *health.Config     `yaml:"health" json:"health" toml:"health"`
	Telemetry        *telemetry.Config  `yaml:"telemetry" json:"telemetry" toml:"telemetry"`
}

// FromConfig builds an App from framework config using the bootstrap assembly pipeline.
func FromConfig(cfg *config.Config, opts ...BootstrapOption) (*App, error) {
	b, err := Assemble(cfg, opts...)
	if err != nil {
		return nil, err
	}
	return NewFromBootstrap(b), nil
}

// Load reads configPath and builds an App from framework config.
func Load(configPath string, opts ...BootstrapOption) (*App, error) {
	if configPath == "" {
		return nil, fmt.Errorf("app: config path is required")
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	return FromConfig(cfg, opts...)
}

// Assemble builds framework runtimes from config without creating the App yet.
func Assemble(cfg *config.Config, opts ...BootstrapOption) (_ *Bootstrap, retErr error) {
	if cfg == nil {
		return nil, fmt.Errorf("app: config cannot be nil")
	}

	var bcfg BootstrapConfig
	for _, opt := range opts {
		if opt != nil {
			opt(&bcfg)
		}
	}

	fc, err := decodeFrameworkConfig(cfg)
	if err != nil {
		return nil, err
	}

	var rb rollbackStack
	defer func() {
		if retErr != nil {
			rb.Rollback()
		}
	}()

	log, err := xlog.New(&fc.Logger)
	if err != nil {
		return nil, fmt.Errorf("app: logger: %w", err)
	}
	rb.Add(log.Close)

	hrt := health.NewRuntime(fc.Health)
	telemetryRuntime, err := telemetry.New(fc.Telemetry, bcfg.TelemetryOptions...)
	if err != nil {
		return nil, fmt.Errorf("app: telemetry: %w", err)
	}
	rb.Add(telemetryRuntime.Close)

	rpcOpts := append([]rpc.Option{}, bcfg.RPCServerOptions...)
	if telemetryRuntime != nil && telemetryRuntime.Trace != nil && telemetryRuntime.Trace.Enabled() {
		rpcOpts = append(rpcOpts, rpc.WithStatsHandlers(obstrace.ServerHandler()))
	}
	if telemetryRuntime != nil && telemetryRuntime.MetricsPath != "" {
		rpcOpts = append(rpcOpts,
			rpc.WithUnaryInterceptors(obsmetrics.UnaryServerInterceptor()),
			rpc.WithStreamInterceptors(obsmetrics.StreamServerInterceptor()),
		)
	}

	rpcRuntime, err := rpc.NewRuntime(log, &rpc.RuntimeConfig{
		Server:        fc.RPCServer,
		ClientOptions: fc.RPCClientOptions,
		Etcd:          fc.Etcd,
	}, rpcOpts...)
	if err != nil {
		return nil, fmt.Errorf("app: rpc runtime: %w", err)
	}
	rb.Add(rpcRuntime.Close)

	var apiRuntime APIRuntime
	if fc.APIServer != nil {
		apiOpts := append([]api.Option{}, bcfg.APIOptions...)
		if telemetryRuntime != nil && telemetryRuntime.MetricsPath != "" {
			apiOpts = append(apiOpts, api.WithMiddleware(obsmetrics.GinMiddleware()))
		}
		if telemetryRuntime != nil && telemetryRuntime.Trace != nil && telemetryRuntime.Trace.Enabled() {
			apiOpts = append(apiOpts, api.WithMiddleware(obstrace.Gin(telemetryRuntime.ServiceName)))
		}
		apiRuntime, err = api.NewServer(log, fc.APIServer, apiOpts...)
		if err != nil {
			return nil, fmt.Errorf("app: api runtime: %w", err)
		}
		if hrt != nil && hrt.Registry != nil && hrt.Path != "" {
			if err := apiRuntime.Register(func(engine *api.Engine) {
				engine.GET(hrt.Path, health.Handler(hrt.Registry))
			}); err != nil {
				return nil, fmt.Errorf("app: register health endpoint: %w", err)
			}
		}
		if telemetryRuntime != nil && telemetryRuntime.MetricsPath != "" {
			if err := apiRuntime.Register(func(engine *api.Engine) {
				engine.GET(telemetryRuntime.MetricsPath, func(c *gin.Context) {
					obsmetrics.Handler().ServeHTTP(c.Writer, c.Request)
				})
			}); err != nil {
				return nil, fmt.Errorf("app: register metrics endpoint: %w", err)
			}
		}
	}

	var resourceRuntime ResourceRuntime
	if fc.Resources != nil {
		resourceRuntime, err = resource.New(fc.Resources, bcfg.ResourceOptions...)
		if err != nil {
			return nil, fmt.Errorf("app: resource runtime: %w", err)
		}
		rb.Add(resourceRuntime.Close)
	}

	var shutdownTimeout time.Duration
	if rawTimeout, ok := cfg.Get("shutdownTimeout"); ok {
		shutdownTimeout, err = parseShutdownTimeout(rawTimeout)
		if err != nil {
			return nil, err
		}
	}

	return &Bootstrap{
		Logger:          log,
		API:             apiRuntime,
		RPC:             rpcRuntime,
		Jobs:            job.NewScheduler(log),
		Resources:       resourceRuntime,
		Telemetry:       telemetryRuntime,
		Health:          hrt,
		ShutdownTimeout: shutdownTimeout,
	}, nil
}

// NewFromBootstrap creates an App from already assembled framework runtimes.
func NewFromBootstrap(b *Bootstrap, opts ...Option) *App {
	if b == nil {
		return New(nil, opts...)
	}

	baseOpts := make([]Option, 0, 5+len(opts))
	if b.RPC != nil {
		baseOpts = append(baseOpts, WithRPCRuntime(b.RPC))
	}
	if b.API != nil {
		baseOpts = append(baseOpts, WithAPIRuntime(b.API))
	}
	if b.Jobs != nil {
		baseOpts = append(baseOpts, WithJobRuntime(b.Jobs))
	}
	if b.Resources != nil {
		baseOpts = append(baseOpts, WithResourceRuntime(b.Resources))
	}
	if b.Telemetry != nil {
		baseOpts = append(baseOpts, WithTelemetry(b.Telemetry))
	}
	if b.ShutdownTimeout > 0 {
		baseOpts = append(baseOpts, WithShutdownTimeout(b.ShutdownTimeout))
	}
	baseOpts = append(baseOpts, opts...)
	return New(b.Logger, baseOpts...)
}

// parseShutdownTimeout converts supported config values into a duration.
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

// decodeFrameworkConfig decodes framework-owned config sections from cfg.
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
	if err := decodeKey("health", &fc.Health); err != nil {
		return nil, err
	}
	if err := decodeKey("telemetry", &fc.Telemetry); err != nil {
		return nil, err
	}

	return fc, nil
}
