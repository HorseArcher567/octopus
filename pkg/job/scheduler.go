package job

import (
	"context"
	"sync"

	"github.com/HorseArcher567/octopus/pkg/xlog"
)

type Scheduler struct {
	log    *xlog.Logger
	jobs   []*Job
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewScheduler(log *xlog.Logger) *Scheduler {
	return &Scheduler{
		log:  log,
		jobs: make([]*Job, 0),
	}
}

func (s *Scheduler) AddJob(job *Job) {
	s.jobs = append(s.jobs, job)
}

// Start starts all jobs in background goroutines and returns immediately.
// Use Stop to gracefully shut down the scheduler and wait for all jobs to finish.
func (s *Scheduler) Start() error {
	// Create context for job lifecycle management
	s.ctx, s.cancel = context.WithCancel(context.Background())

	s.log.Info("starting job scheduler", "jobCount", len(s.jobs))

	// Start all jobs in background
	for _, job := range s.jobs {
		job := job
		s.wg.Go(func() {
			if err := job.Run(s.ctx, s.log); err != nil {
				s.log.Error("job run failed", "name", job.Name, "error", err)
			}
		})
	}

	return nil
}

// Stop gracefully stops the scheduler by cancelling the context and waiting for all jobs to finish.
// It uses the given context for timeout control. If the timeout is reached, it will return an error
// but jobs may still be running in the background.
func (s *Scheduler) Stop(ctx context.Context) error {
	s.log.Info("shutting down job scheduler gracefully")

	// Cancel context to signal all jobs to stop
	s.cancel()

	// Wait for all jobs to finish with timeout
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.log.Info("all jobs finished, scheduler stopped")
		return nil
	case <-ctx.Done():
		s.log.Warn("job scheduler shutdown timeout, some jobs may still be running")
		return ctx.Err()
	}
}
