package service

import (
	"context"
	"github.com/k8s-practice/octopus/pkg/log"
)

var (
	registeredEntries = make(map[string]Entry)
)

// Entry define service interfaces.
type Entry interface {
	Enabled() bool
	Name() string
	Serve(ctx context.Context)
	Stop(ctx context.Context)
}

func registerEntry(entries ...Entry) {
	for i := 0; i < len(entries); i++ {
		if _, ok := registeredEntries[entries[i].Name()]; ok {
			log.Panicf("repeated registerEntry %s service", entries[i].Name())
		} else {
			registeredEntries[entries[i].Name()] = entries[i]
		}
	}
}
