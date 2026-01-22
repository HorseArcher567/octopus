package job

import (
	"context"
	"errors"

	"github.com/HorseArcher567/octopus/pkg/xlog"
)

type Func func(ctx context.Context, log *xlog.Logger) error

type Job struct {
	// Job name
	Name string `yaml:"name" json:"name" toml:"name"`
	// Job function
	Func Func `yaml:"func" json:"func" toml:"func"`
}

func (j *Job) Validate() error {
	if j.Name == "" {
		return errors.New("job name is required")
	}

	if j.Func == nil {
		return errors.New("job function is required")
	}

	return nil
}

func (j *Job) Run(ctx context.Context, log *xlog.Logger) error {
	log.Info("running job", "name", j.Name)
	return j.Func(ctx, log)
}
