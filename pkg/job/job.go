package job

import (
	"context"
	"errors"

	"github.com/HorseArcher567/octopus/pkg/store"
	"github.com/HorseArcher567/octopus/pkg/xlog"
)

type Func func(*Context) error

type SchedulerConfig struct {
	Logger string `yaml:"logger" json:"logger" toml:"logger"`
}

type Context struct {
	store.Reader
	ctx  context.Context
	log  *xlog.Logger
	name string
}

func NewContext(ctx context.Context, log *xlog.Logger, reader store.Reader, name string) *Context {
	return &Context{Reader: reader, ctx: ctx, log: log, name: name}
}

func (c *Context) Context() context.Context { return c.ctx }

func (c *Context) Logger() *xlog.Logger { return c.log }

func (c *Context) Name() string { return c.name }

type Job struct {
	Name string `yaml:"name" json:"name" toml:"name"`
	Func Func   `yaml:"func" json:"func" toml:"func"`
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

func (j *Job) Run(ctx *Context) error {
	ctx.Logger().Info("running job", "name", j.Name)
	return j.Func(ctx)
}
