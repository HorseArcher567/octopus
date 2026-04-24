package hook

import (
	"context"

	"github.com/HorseArcher567/octopus/pkg/store"
	"github.com/HorseArcher567/octopus/pkg/xlog"
)

type Func func(*Context) error

type Context struct {
	store.Reader
	ctx context.Context
	log *xlog.Logger
}

func NewContext(ctx context.Context, log *xlog.Logger, reader store.Reader) *Context {
	return &Context{Reader: reader, ctx: ctx, log: log}
}

func (c *Context) Context() context.Context { return c.ctx }

func (c *Context) Logger() *xlog.Logger { return c.log }
