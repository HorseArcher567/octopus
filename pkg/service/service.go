package service

import (
	"context"
	"github.com/k8s-practice/octopus/pkg/log"
	"sync"
)

// Init inits all services.
func Init(bootConfig map[interface{}]interface{}) {
	registerEntry(buildEntry(bootConfig)...)
}

// Start starts all services by invoke Entry.Serve.
func Start(ctx context.Context) {
	for _, entry := range registeredEntries {
		if entry.Enabled() {
			go entry.Serve(ctx)
		} else {
			log.Warnln(entry.Name(), "not enabled.")
		}
	}
}

// Stop stops all services by invoke Entry.Stop.
func Stop(ctx context.Context) {
	var wg sync.WaitGroup
	defer wg.Wait()

	for _, entry := range registeredEntries {
		if !entry.Enabled() {
			continue
		}

		wg.Add(1)
		go func(entry Entry) {
			defer wg.Done()
			entry.Stop(ctx)
		}(entry)
	}
}
