package octopus

import (
	"github.com/k8s-practice/octopus/config"
	"github.com/mitchellh/mapstructure"
)

func New() *Octopus {
	return &Octopus{
		frameInit: make([]func() error, 0),
		appInit:   make([]func() error, 0),
	}

}

// Brain is the brain of octopus.
type Octopus struct {
	frameLogger interface{}

	conf config.Config

	// frameInit is the framework initialize functions slice.
	// It's will be invoked first.
	frameInit []func() error

	// appInit is the application initialize functions slice.
	// It's will be invoded after Run function.
	appInit []func() error
}

// WithConfig sets Octopus.conf .
func (o *Octopus) WithConfig(c config.Config) *Octopus {
	o.conf = c

	return o
}

// Load loads configuration by key from Octopus.conf .
func (o *Octopus) Load(key string, i interface{}) error {
	err := mapstructure.Decode(o.conf.Get(key), i)
	return err
}
