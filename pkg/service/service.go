package service

import (
	"context"
	"github.com/k8s-practice/octopus/pkg/log"
	"time"
)

type Entry interface {
	Name() string
	Run(ctx context.Context)
	Stop(ctx context.Context)
}

var (
	registeredEntries  = make(map[string]Entry)
	registeredBuilders []Builder
)

func Register(entries ...Entry) {
	for i := 0; i < len(entries); i++ {
		if _, ok := registeredEntries[entries[i].Name()]; ok {
			log.Panicf("repeated register %s service", entries[i].Name())
		} else {
			registeredEntries[entries[i].Name()] = entries[i]
		}
	}
}

func Get(name string) Entry {
	if service, ok := registeredEntries[name]; ok {
		return service
	} else {
		return nil
	}
}

type Builder interface {
	Build(bootConfig map[interface{}]interface{}) []Entry
}

func RegisterBuilder(builder Builder) {
	registeredBuilders = append(registeredBuilders, builder)
}

func BuildEntries(bootConfig map[interface{}]interface{}) {
	for _, builder := range registeredBuilders {
		services := builder.Build(bootConfig)
		Register(services...)
	}
}

func Run() {
	ctx := context.Background()
	for _, entry := range registeredEntries {
		go entry.Run(ctx)
	}
}

func Stop() {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*3)
	for _, entry := range registeredEntries {
		entry.Stop(ctx)
	}
}
