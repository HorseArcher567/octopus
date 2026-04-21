package assemble

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/HorseArcher567/octopus/pkg/app"
	"github.com/HorseArcher567/octopus/pkg/config"
	"github.com/HorseArcher567/octopus/pkg/database"
	"github.com/HorseArcher567/octopus/pkg/etcd"
	"github.com/HorseArcher567/octopus/pkg/job"
	redisclient "github.com/HorseArcher567/octopus/pkg/redis"
	"github.com/HorseArcher567/octopus/pkg/rpc"
	"github.com/HorseArcher567/octopus/pkg/store"
	"github.com/HorseArcher567/octopus/pkg/xlog"
)

type assembleConfig struct {
	App app.Config `yaml:"app" json:"app" toml:"app"`

	Logger []xlog.Config        `yaml:"logger" json:"logger" toml:"logger"`
	Etcd   []etcd.Config        `yaml:"etcd" json:"etcd" toml:"etcd"`
	MySQL  []database.Config    `yaml:"mysql" json:"mysql" toml:"mysql"`
	Redis  []redisclient.Config `yaml:"redis" json:"redis" toml:"redis"`

	APIServer        *api.ServerConfig    `yaml:"apiServer" json:"apiServer" toml:"apiServer"`
	RPCServer        *rpc.ServerConfig    `yaml:"rpcServer" json:"rpcServer" toml:"rpcServer"`
	JobScheduler     *job.SchedulerConfig `yaml:"jobScheduler" json:"jobScheduler" toml:"jobScheduler"`
	RPCClientOptions *rpc.ClientOptions   `yaml:"rpcClientOptions" json:"rpcClientOptions" toml:"rpcClientOptions"`
}

type state struct {
	cfg   *assembleConfig
	log   *xlog.Logger
	store store.Store

	api apiServer
	rpc rpcServer
	job jobScheduler
}

type setupStep struct {
	name string
	run  func(*state) error
}

var defaultSetupSteps = []setupStep{
	{name: "loggers", run: setupLoggersStep},
	{name: "app-logger", run: selectAppLoggerStep},
	{name: "rpc-client-options", run: setupRPCClientOptionsStep},
	{name: "jobs", run: setupJobsStep},
	{name: "etcd", run: setupEtcdStep},
	{name: "mysql", run: setupMySQLStep},
	{name: "redis", run: setupRedisStep},
	{name: "api", run: setupAPIStep},
	{name: "rpc", run: setupRPCStep},
}

func setup(raw *config.Config) (_ *state, retErr error) {
	cfg, err := decodeAssembleConfig(raw)
	if err != nil {
		return nil, err
	}

	s := &state{
		cfg:   cfg,
		store: store.New(),
	}
	defer func() {
		if retErr != nil {
			_ = s.store.Close()
		}
	}()

	if err := runSetupSteps(s, defaultSetupSteps); err != nil {
		return nil, err
	}
	return s, nil
}

func decodeAssembleConfig(raw *config.Config) (*assembleConfig, error) {
	if raw == nil {
		return nil, fmt.Errorf("assemble: config cannot be nil")
	}
	cfg := &assembleConfig{}
	if err := raw.UnmarshalStrict(cfg); err != nil {
		return nil, fmt.Errorf("assemble: invalid config: %w", err)
	}
	return cfg, nil
}

func runSetupSteps(s *state, steps []setupStep) error {
	for _, step := range steps {
		if err := step.run(s); err != nil {
			return fmt.Errorf("assemble: setup %s: %w", step.name, err)
		}
	}
	return nil
}

func setupLoggersStep(s *state) error {
	items := s.cfg.Logger
	if len(items) == 0 {
		return fmt.Errorf("assemble: logger is required")
	}

	seen := make(map[string]struct{}, len(items))
	for i, item := range items {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			return fmt.Errorf("assemble: logger[%d]: name is required", i)
		}
		if _, ok := seen[name]; ok {
			return fmt.Errorf("assemble: logger[%s]: duplicate name", name)
		}
		seen[name] = struct{}{}

		log, err := xlog.New(&item)
		if err != nil {
			return fmt.Errorf("assemble: logger[%s]: %w", name, err)
		}
		if err := s.store.SetNamed(name, log, store.WithClose(log.Close)); err != nil {
			_ = log.Close()
			return fmt.Errorf("assemble: logger[%s]: %w", name, err)
		}
	}
	return nil
}

func selectAppLoggerStep(s *state) error {
	selected := strings.TrimSpace(s.cfg.App.Logger)
	if selected == "" {
		return fmt.Errorf("assemble: app.logger is required")
	}

	log, err := lookupLogger(selected, s.store)
	if err != nil {
		return fmt.Errorf("assemble: app.logger: %w", err)
	}
	s.log = log
	return nil
}

func selectComponentLogger(name string, fallback *xlog.Logger, st store.Store) (*xlog.Logger, error) {
	selected := strings.TrimSpace(name)
	if selected == "" {
		if fallback == nil {
			return nil, fmt.Errorf("logger fallback cannot be nil")
		}
		return fallback, nil
	}
	return lookupLogger(selected, st)
}

func lookupLogger(name string, st store.Store) (*xlog.Logger, error) {
	value, err := st.GetNamed(name, reflect.TypeFor[*xlog.Logger]())
	if err != nil {
		return nil, fmt.Errorf("logger %q not found", name)
	}
	log, ok := value.(*xlog.Logger)
	if !ok || log == nil {
		return nil, fmt.Errorf("logger %q has invalid type %T", name, value)
	}
	return log, nil
}

func setupRPCClientOptionsStep(s *state) error {
	if s.cfg.RPCClientOptions == nil {
		return nil
	}
	if err := s.store.SetNamed("default", s.cfg.RPCClientOptions); err != nil {
		return fmt.Errorf("assemble: rpcClientOptions[default]: %w", err)
	}
	return nil
}

func setupJobsStep(s *state) error {
	loggerName := ""
	if s.cfg.JobScheduler != nil {
		loggerName = s.cfg.JobScheduler.Logger
	}
	log, err := selectComponentLogger(loggerName, s.log, s.store)
	if err != nil {
		return fmt.Errorf("assemble: jobScheduler.logger: %w", err)
	}
	s.job = job.NewScheduler(log)
	return nil
}

func setupEtcdStep(s *state) error {
	items := s.cfg.Etcd
	if len(items) == 0 {
		return nil
	}
	if err := validateEtcdConfigs(items); err != nil {
		return err
	}
	for _, item := range items {
		client, err := etcd.NewClient(&item)
		if err != nil {
			return fmt.Errorf("assemble: etcd[%s]: %w", item.Name, err)
		}
		if err := s.store.SetNamed(item.Name, client, store.WithClose(client.Close)); err != nil {
			_ = client.Close()
			return fmt.Errorf("assemble: etcd[%s]: %w", item.Name, err)
		}
	}
	return nil
}

func setupMySQLStep(s *state) error {
	items := s.cfg.MySQL
	if len(items) == 0 {
		return nil
	}
	if err := validateMySQLConfigs(items); err != nil {
		return err
	}
	for _, item := range items {
		db, err := database.New(&item)
		if err != nil {
			return fmt.Errorf("assemble: mysql[%s]: %w", item.Name, err)
		}
		if err := db.PingTimeout(item.PingTimeout); err != nil {
			_ = db.Close()
			return fmt.Errorf("assemble: mysql[%s]: ping failed: %w", item.Name, err)
		}
		if err := s.store.SetNamed(item.Name, db, store.WithClose(db.Close)); err != nil {
			_ = db.Close()
			return fmt.Errorf("assemble: mysql[%s]: %w", item.Name, err)
		}
	}
	return nil
}

func setupRedisStep(s *state) error {
	items := s.cfg.Redis
	if len(items) == 0 {
		return nil
	}
	if err := validateRedisConfigs(items); err != nil {
		return err
	}
	for _, item := range items {
		client, err := redisclient.New(&item)
		if err != nil {
			return fmt.Errorf("assemble: redis[%s]: %w", item.Name, err)
		}
		if err := client.PingTimeout(item.PingTimeout); err != nil {
			_ = client.Close()
			return fmt.Errorf("assemble: redis[%s]: ping failed: %w", item.Name, err)
		}
		if err := s.store.SetNamed(item.Name, client, store.WithClose(client.Close)); err != nil {
			_ = client.Close()
			return fmt.Errorf("assemble: redis[%s]: %w", item.Name, err)
		}
	}
	return nil
}

func setupAPIStep(s *state) error {
	if s.cfg.APIServer == nil {
		return nil
	}
	log, err := selectComponentLogger(s.cfg.APIServer.Logger, s.log, s.store)
	if err != nil {
		return fmt.Errorf("assemble: apiServer.logger: %w", err)
	}
	server, err := api.NewServer(log, s.cfg.APIServer)
	if err != nil {
		return fmt.Errorf("assemble: api server: %w", err)
	}
	s.api = server
	return nil
}

func setupRPCStep(s *state) error {
	if s.cfg.RPCServer == nil {
		return nil
	}
	log, err := selectComponentLogger(s.cfg.RPCServer.Logger, s.log, s.store)
	if err != nil {
		return fmt.Errorf("assemble: rpcServer.logger: %w", err)
	}
	server, err := rpc.NewServer(log, s.cfg.RPCServer)
	if err != nil {
		return fmt.Errorf("assemble: rpc server: %w", err)
	}
	s.rpc = server
	return nil
}

func validateEtcdConfigs(items []etcd.Config) error {
	seen := make(map[string]struct{}, len(items))
	for i, item := range items {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			return fmt.Errorf("assemble: etcd[%d]: name is required", i)
		}
		if _, ok := seen[name]; ok {
			return fmt.Errorf("assemble: etcd[%s]: duplicate name", name)
		}
		seen[name] = struct{}{}
		if err := item.Validate(); err != nil {
			return fmt.Errorf("assemble: etcd[%s]: %w", name, err)
		}
	}
	return nil
}

func validateMySQLConfigs(items []database.Config) error {
	seen := make(map[string]struct{}, len(items))
	for i, item := range items {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			return fmt.Errorf("assemble: mysql[%d]: name is required", i)
		}
		if _, ok := seen[name]; ok {
			return fmt.Errorf("assemble: mysql[%s]: duplicate name", name)
		}
		seen[name] = struct{}{}
		if err := item.Validate(); err != nil {
			return fmt.Errorf("assemble: mysql[%s]: %w", name, err)
		}
	}
	return nil
}

func validateRedisConfigs(items []redisclient.Config) error {
	seen := make(map[string]struct{}, len(items))
	for i, item := range items {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			return fmt.Errorf("assemble: redis[%d]: name is required", i)
		}
		if _, ok := seen[name]; ok {
			return fmt.Errorf("assemble: redis[%s]: duplicate name", name)
		}
		seen[name] = struct{}{}
		if err := item.Validate(); err != nil {
			return fmt.Errorf("assemble: redis[%s]: %w", name, err)
		}
	}
	return nil
}
