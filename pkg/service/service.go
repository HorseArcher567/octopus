package service

import (
	"context"
)

func Init(bootConfig map[interface{}]interface{}) {
	buildEntries(bootConfig)
}

func Run(ctx context.Context) {
	for _, entry := range registeredEntries {
		go entry.Run(ctx)
	}
}

func Stop(ctx context.Context) {
	for _, entry := range registeredEntries {
		entry.Stop(ctx)
	}
}
