package service

import (
	"context"
	"sync"
)

// Init init all services.
func Init(bootConfig map[interface{}]interface{}) {
	registerEntry(buildEntry(bootConfig)...)
}

// Start start all services by invoke Entry.Serve.
func Start(ctx context.Context) {
	for _, entry := range registeredEntries {
		go entry.Serve(ctx)
	}
}

// Stop stop all services by invoke Entry.Stop.
func Stop(ctx context.Context) {
	var wg sync.WaitGroup
	defer wg.Wait()

	for _, entry := range registeredEntries {
		wg.Add(1)
		go func(entry Entry) {
			defer wg.Done()
			entry.Stop(ctx)
		}(entry)
	}
}
