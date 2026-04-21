package assemble

import (
	"context"
	"errors"
	"fmt"

	"github.com/HorseArcher567/octopus/pkg/app"
)

var (
	ErrAPINotConfigured = errors.New("assemble: api not configured")
	ErrRPCNotConfigured = errors.New("assemble: rpc not configured")
)

type namedService struct {
	name string
	run  func(context.Context) error
	stop func(context.Context) error
}

func (s *namedService) Name() string { return s.name }

func (s *namedService) Run(ctx context.Context) error {
	name := "<nil>"
	if s != nil {
		name = s.name
	}
	if s == nil || s.run == nil {
		return fmt.Errorf("assemble: service %q has no run function", name)
	}
	return s.run(ctx)
}

func (s *namedService) Stop(ctx context.Context) error {
	name := "<nil>"
	if s != nil {
		name = s.name
	}
	if s == nil || s.stop == nil {
		return fmt.Errorf("assemble: service %q has no stop function", name)
	}
	return s.stop(ctx)
}

func builtinServices(s *state) []app.Service {
	services := make([]app.Service, 0, 3)
	if s.api != nil {
		services = append(services, &namedService{name: "api", run: s.api.Run, stop: s.api.Stop})
	}
	if s.rpc != nil {
		services = append(services, &namedService{name: "rpc", run: s.rpc.Run, stop: s.rpc.Stop})
	}
	if s.job != nil {
		services = append(services, &namedService{name: "jobs", run: s.job.Run, stop: s.job.Stop})
	}
	return services
}
