package job

import (
	"context"

	"github.com/HorseArcher567/octopus/pkg/store"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	"golang.org/x/sync/errgroup"
)

type Scheduler struct {
	log    *xlog.Logger
	jobs   []*Job
	g      *errgroup.Group
	ctx    context.Context
	cancel context.CancelFunc
	store  store.Reader
}

func NewScheduler(log *xlog.Logger, reader store.Reader) *Scheduler {
	return &Scheduler{
		log:   log,
		jobs:  make([]*Job, 0),
		store: reader,
	}
}

func (s *Scheduler) Add(name string, fn Func) error {
	job := &Job{Name: name, Func: fn}
	if err := job.Validate(); err != nil {
		return err
	}
	s.jobs = append(s.jobs, job)
	return nil
}

func (s *Scheduler) AddJob(job *Job) {
	if job == nil {
		return
	}
	s.jobs = append(s.jobs, job)
}

// HasJobs returns true if there are registered jobs.
func (s *Scheduler) HasJobs() bool {
	return len(s.jobs) > 0
}

// Run starts all jobs and blocks until ctx is cancelled or all jobs complete.
// If no jobs are registered, returns nil immediately (does not block).
func (s *Scheduler) Run(ctx context.Context) error {
	if len(s.jobs) == 0 {
		s.log.Debug("no jobs registered, scheduler exiting")
		return nil // No jobs, return immediately, do not block
	}

	s.ctx, s.cancel = context.WithCancel(ctx)
	s.g, _ = errgroup.WithContext(s.ctx)

	s.log.Info("starting job scheduler", "jobCount", len(s.jobs))

	for _, job := range s.jobs {
		job := job
		s.g.Go(func() error {
			return job.Run(NewContext(s.ctx, s.log, s.store, job.Name))
		})
	}

	return s.g.Wait()
}

// Stop stops all jobs by cancelling the context.
// The caller (App) is responsible for calling this method.
// g.Wait() is called in Run(), so we don't need to wait here.
func (s *Scheduler) Stop(ctx context.Context) error {
	s.log.Info("stopping job scheduler")

	// If Run has never been called (no jobs, or scheduler not started),
	// s.ctx will be nil and there's nothing to wait for.
	if s.ctx == nil {
		return nil
	}

	if s.cancel != nil {
		s.cancel()
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.ctx.Done():
		return nil
	}
}
