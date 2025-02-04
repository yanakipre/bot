package scheduletooling

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/yanakipre/bot/internal/logger"
)

// updateSchedules returns function that keeps track of schedules and updates the triggers
func (s *Scheduler) updateSchedules(ctx context.Context) func(context.Context) {
	s.jobsMutex.Lock()
	oldConfigs := map[int]Config{}
	for _, job := range s.jobs {
		newConfig, err := job.GetConfig()
		if err != nil {
			logger.Fatal(
				ctx,
				"could not get config, skipping job",
				jobLogKeys(job, zap.Error(err))...)
		}
		logger.Debug(ctx, "remembered", jobLogKeys(job)...)
		oldConfigs[job.Key()] = newConfig
	}
	s.jobsMutex.Unlock()

	return func(ctx context.Context) {
		timer := time.NewTimer(s.jobUpdateInterval)
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				// just have a function so that the
				// defers fire at the correct times
				func() {
					defer timer.Reset(s.jobUpdateInterval)

					s.jobsMutex.Lock()
					defer s.jobsMutex.Unlock()

					for _, job := range s.jobs {
						newConfig, err := job.GetConfig()
						if err != nil {
							logger.Error(
								ctx,
								"could not get config, skipping job",
								jobLogKeys(job, zap.Error(err))...)
						}
						oldConfig, ok := oldConfigs[job.Key()]
						oldConfigs[job.Key()] = newConfig
						if !ok {
							// apply all settings, because it might have already executed with another
							// config
							job.SetTrigger(newConfig.GetTriggerValidated())
							if newConfig.Enabled {
								s.scheduleSilently(ctx, job)
								logger.Info(ctx, "scheduled",
									jobLogKeys(
										job,
										zap.Bool("enabled", newConfig.Enabled),
										zap.String(
											"trigger",
											newConfig.GetTriggerValidated().Description(),
										),
									)...)
							} else {
								s.stopScheduling(ctx, job)
							}
							continue
						}
						if oldConfig == newConfig {
							// nothing changed
							continue
						}
						if oldConfig.Enabled != newConfig.Enabled {
							// turn on or off
							if newConfig.Enabled {
								s.scheduleSilently(ctx, job)
								logger.Info(ctx, "scheduled",
									jobLogKeys(
										job,
										zap.Bool("enabled", newConfig.Enabled),
										zap.String(
											"trigger",
											newConfig.GetTriggerValidated().Description(),
										),
									)...)
							} else {
								s.stopScheduling(ctx, job)
							}
							continue
						}
						if oldConfig.Interval != newConfig.Interval {
							logger.Info(
								ctx,
								"interval changed",
								jobLogKeys(
									job,
									zap.String(
										"trigger",
										newConfig.GetTriggerValidated().Description(),
									),
								)...)
							// it will be applied when next scheduling calculation occurs. This is not an
							// immediate action.
							job.SetTrigger(newConfig.GetTriggerValidated())
							continue
						}
						logger.Warn(
							ctx,
							fmt.Sprintf(
								"unsupported change for configs, old: %+v, new: %+v",
								oldConfig,
								newConfig,
							),
						)
					}
				}() // function called in the loop
			}
		}
	}
}
