package scheduletooling

import (
	"errors"
	"fmt"

	"github.com/reugn/go-quartz/quartz"
	"go.uber.org/zap"

	"github.com/yanakipre/bot/internal/clouderr"
	"github.com/yanakipre/bot/internal/encodingtooling"
)

// Config declares options supported by cronjob.
type Config struct {
	// UniqueName is an identifier unique for scheduler
	// The uniqueness of a job is a limitation of the underlying quartz library.
	// not expected to be configurable via config files.
	UniqueName string `yaml:"-"                         json:"-"`
	// Enabled false job means should not run
	//
	// Should be FALSE by default, until #1644 is resolved.
	// This is a safety measure to prevent accidental
	// concurrent access from both control plane and Console.
	Enabled bool `yaml:"enabled"                   json:"enabled"`
	// Interval sets interval this job will run at
	Interval encodingtooling.Duration `yaml:"interval"                  json:"interval"`
	// CronExpression sets cron expression this job will run at
	// It has precedence over Interval
	CronExpression string `yaml:"cron_expression,omitempty" json:"cron_expression,omitempty"`
	// RunOnStart is supported for yanakipreriver.PeriodicJob only.
	//
	// RunOnStart can be used to indicate that a periodic job should run an
	// initial job as a new scheduler is started. This can be used as a hedge
	// for jobs with longer scheduled durations that may not get to expiry
	// before a new scheduler is elected.
	RunOnStart bool `yaml:"run_on_start,omitempty"`
	// Timeout is supported for yanakipreriver.PeriodicJob only.
	//
	// If Timeout has elapsed, then the river job rescuer
	// retries or discards the job, based on number of retries left.
	Timeout encodingtooling.Duration `yaml:"timeout"`
}

func (c *Config) Validate() error {
	if c.Enabled {
		if c.Interval.Duration == 0 && c.CronExpression == "" {
			return errors.New("interval or cron_expression must be set")
		}
	}

	if c.CronExpression != "" {
		_, err := quartz.NewCronTrigger(c.CronExpression)
		if err != nil {
			return fmt.Errorf("error validating cron expression: %w",
				clouderr.WrapWithFields(err, zap.String("cron_expression", c.CronExpression)),
			)
		}
	}

	return nil
}

// GetTriggerValidated returns a quartz trigger based on the config.
// It assumes that the config has been validated.
func (c *Config) GetTriggerValidated() quartz.Trigger {
	if c.CronExpression != "" {
		trigger, _ := quartz.NewCronTrigger(c.CronExpression)
		return trigger
	}
	return quartz.NewSimpleTrigger(c.Interval.Duration)
}
