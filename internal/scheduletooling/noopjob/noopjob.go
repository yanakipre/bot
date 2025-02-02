package noopjob

import (
	"context"
	"fmt"

	"github.com/reugn/go-quartz/quartz"

	"github.com/yanakipe/bot/internal/scheduletooling"
)

type NoopJobImpl struct {
	key int
}

func (n *NoopJobImpl) Execute(_ context.Context)     {}
func (n *NoopJobImpl) Close(_ context.Context) error { return nil }
func (n *NoopJobImpl) Description() string           { return "noop job" }
func (n *NoopJobImpl) Key() int                      { return n.key }

func (n *NoopJobImpl) NextFireTime(prev int64) (int64, error) {
	return prev, fmt.Errorf("noop job has no next fire time")
}

func (n *NoopJobImpl) GetConfig() (scheduletooling.Config, error) {
	return scheduletooling.Config{}, fmt.Errorf("noop job has no config")
}

func (n *NoopJobImpl) SetTrigger(_ quartz.Trigger) {}

var _ scheduletooling.Job = &NoopJobImpl{}

func noopJob() func() scheduletooling.Job {
	cnt := 0
	return func() scheduletooling.Job {
		return &NoopJobImpl{
			key: cnt,
		}
	}
}

var NewNoopJob = noopJob()
