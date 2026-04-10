package health

import (
	"context"
	"sync"
)

// Status represents the aggregate or per-check health state.
type Status string

const (
	StatusUp   Status = "UP"
	StatusDown Status = "DOWN"
)

// Checker performs a health check.
type Checker interface {
	Name() string
	Check(ctx context.Context) error
}

// Detail describes the result of a single checker.
type Detail struct {
	Status Status `json:"status"`
	Error  string `json:"error,omitempty"`
}

// Report contains the aggregate health state and all checker details.
type Report struct {
	Status  Status            `json:"status"`
	Details map[string]Detail `json:"details"`
}

// Registry stores health checkers and can execute them safely.
type Registry struct {
	mu       sync.RWMutex
	checkers map[string]Checker
}

// New creates an empty health registry.
func New() *Registry {
	return &Registry{checkers: make(map[string]Checker)}
}

// Register adds or replaces a named checker.
func (r *Registry) Register(c Checker) error {
	if c == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checkers[c.Name()] = c
	return nil
}

// Check executes all registered health checks and aggregates the result.
func (r *Registry) Check(ctx context.Context) Report {
	r.mu.RLock()
	checkers := make([]Checker, 0, len(r.checkers))
	for _, c := range r.checkers {
		checkers = append(checkers, c)
	}
	r.mu.RUnlock()

	report := Report{
		Status:  StatusUp,
		Details: make(map[string]Detail, len(checkers)),
	}
	for _, c := range checkers {
		if err := c.Check(ctx); err != nil {
			report.Status = StatusDown
			report.Details[c.Name()] = Detail{Status: StatusDown, Error: err.Error()}
			continue
		}
		report.Details[c.Name()] = Detail{Status: StatusUp}
	}
	return report
}
