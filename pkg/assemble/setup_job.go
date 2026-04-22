package assemble

import (
	"fmt"

	"github.com/HorseArcher567/octopus/pkg/job"
)

func setupJobs(c *setupContext) error {
	var cfg job.SchedulerConfig
	if err := c.decodeStruct("jobScheduler", &cfg); err != nil {
		if _, ok := c.cfg.Get("jobScheduler"); ok {
			return err
		}
	}
	log, err := selectComponentLogger(cfg.Logger, c.state.log, c.state.store)
	if err != nil {
		return fmt.Errorf("assemble: jobScheduler.logger: %w", err)
	}
	c.state.job = job.NewScheduler(log)
	return nil
}
