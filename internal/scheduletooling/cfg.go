package scheduletooling

import (
	"errors"
	"fmt"

	"github.com/reugn/go-quartz/quartz"

	"github.com/yanakipe/bot/internal/encodingtooling"
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
			return fmt.Errorf("error validating cron expression '%s': %w", c.CronExpression, err)
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
