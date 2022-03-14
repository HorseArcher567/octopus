package service

import (
	"context"
	"github.com/k8s-practice/octopus/pkg/log"
)

var (
	registeredEntries = make(map[string]Entry)
)

type Entry interface {
	Name() string
	Run(ctx context.Context)
	Stop(ctx context.Context)
}

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
