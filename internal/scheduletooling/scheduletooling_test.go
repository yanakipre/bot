package scheduletooling

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/yanakipre/bot/internal/encodingtooling"
	"github.com/yanakipre/bot/internal/logger"
	"github.com/yanakipre/bot/internal/testtooling"
)

type metricsCollector struct{}

func (m metricsCollector) JobStarted(_ string) (finished func(error)) {
	return func(err error) {
	}
}

func (m metricsCollector) SkippedJob(_ string) {
}

var _ MetricsCollector = &metricsCollector{}

func TestScheduler_Wait(t *testing.T) {
	testtooling.SetNewGlobalLoggerQuietly()
	type args struct {
		ctx context.Context
		job Job
	}
	tests := []struct {
		name string
		args func() (args, func(t *testing.T))
	}{
		{
			name: "wait calls close",
			args: func() (args, func(t *testing.T)) {
				counter := 0
				job := NewConcurrentInProcessJobWithCloser(
					func(ctx context.Context) error {
						logger.Info(ctx, "exec job")
						return nil
					},
					func(ctx context.Context) error {
						logger.Info(ctx, "closing job")
						counter += 1
						return nil
					},
					Config{Enabled: true, UniqueName: "test-job"}, func() (Config, error) {
						return Config{
							Enabled:  true,
							Interval: encodingtooling.NewDuration(30 * time.Millisecond),
						}, nil
					},
					&metricsCollector{},
					1,
				)
				return args{
						ctx: context.Background(),
						job: job,
					}, func(t *testing.T) {
						require.Equal(t, 1, counter)
					}
			},
		},
		{
			name: "disabled job should not call closer",
			args: func() (args, func(t *testing.T)) {
				counter := 0
				job := NewConcurrentInProcessJobWithCloser(
					func(ctx context.Context) error {
						logger.Info(ctx, "exec job")
						return nil
					},
					func(ctx context.Context) error {
						logger.Info(ctx, "closing job")
						counter += 1
						return nil
					},
					Config{Enabled: false, UniqueName: "test-job"},
					func() (Config, error) {
						return Config{Enabled: false}, nil
					},
					&metricsCollector{},
					1,
				)
				return args{
						ctx: context.Background(),
						job: job,
					}, func(t *testing.T) {
						require.Equal(t, 0, counter)
					}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, finish := tt.args()
			s := NewScheduler(time.Second)
			err := s.Add(args.ctx, args.job)
			require.NoError(t, err)
			s.Start(args.ctx)
			s.Stop()
			s.Wait(args.ctx)
			finish(t)
		})
	}
}
