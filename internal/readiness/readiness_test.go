package readiness

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kamilsk/retry/v5"
	"github.com/kamilsk/retry/v5/strategy"
	"github.com/stretchr/testify/require"

	"github.com/yanakipre/bot/internal/logger"
)

type ready struct{}

func (n *ready) Ready(_ context.Context) error {
	return nil
}

type readyFromNthAttempt struct {
	attemptsGiven int
	calledTimes   int
}

func (n *readyFromNthAttempt) Ready(_ context.Context) error {
	n.calledTimes += 1
	switch {
	case n.calledTimes < n.attemptsGiven:
		return fmt.Errorf("not ready at attempt %d, need %d", n.calledTimes+1, n.attemptsGiven)
	case n.calledTimes == n.attemptsGiven:
		return nil
	default:
		panic(fmt.Sprintf("been called more than %d times", n.attemptsGiven))
	}
}

type neverReady struct {
	err error
}

func (n *neverReady) Ready(_ context.Context) error {
	return n.err
}

func TestReadiness_TryAdd(t *testing.T) {
	type nonImplementing struct{}
	type fields struct {
		checkers []*state
	}
	type args struct {
		candidate any
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		want        bool
		checkersLen int
	}{
		{
			name:   "implementing",
			fields: fields{},
			args: args{
				candidate: &ready{},
			},
			want:        true,
			checkersLen: 1,
		},
		{
			name:   "not implementing",
			fields: fields{},
			args: args{
				candidate: &nonImplementing{},
			},
			want:        false,
			checkersLen: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Readiness{
				checkers: tt.fields.checkers,
			}
			got := r.TryAdd(tt.args.candidate)
			require.Equal(t, tt.want, got)
			require.Equal(t, tt.checkersLen, len(r.checkers))
		})
	}
}

func TestReadiness_IsReady(t *testing.T) {
	type fields struct {
		checkers   []*state
		strategies retry.How
	}
	errToExpect := fmt.Errorf("test-error")
	tests := []struct {
		name   string
		fields fields
		want   error
	}{
		{
			name: "does not check what's ready",
			fields: fields{
				checkers: []*state{
					{
						checker: &readyFromNthAttempt{attemptsGiven: 2},
						name:    "ready from second time",
					},
				},
				strategies: retry.How{
					strategy.Limit(2),
				},
			},
			want: nil,
		},
		{
			name: "not ready exhausted attempts",
			fields: fields{
				checkers: []*state{
					{
						checker: &neverReady{err: errToExpect},
						name:    "never-ready",
					},
				},
				strategies: retry.How{
					strategy.Limit(2),
				},
			},
			want: errToExpect,
		},
		{
			name: "ready",
			fields: fields{
				checkers: []*state{
					{
						checker: &ready{},
						name:    "ready",
					},
				},
			},
			want: nil,
		},
		{
			name: "ready with optional checks",
			fields: fields{
				checkers: []*state{
					{
						checker: &ready{},
						name:    "ready",
					},
					{
						checker:    &neverReady{err: errToExpect},
						name:       "never-ready",
						isOptional: true,
					},
				},
			},
			want: nil,
		},
		{
			name: "not ready with optional checks",
			fields: fields{
				checkers: []*state{
					{
						checker: &neverReady{err: errToExpect},
						name:    "never-ready",
					},
					{
						checker:    &neverReady{err: errToExpect},
						name:       "never-ready",
						isOptional: true,
					},
				},
				strategies: retry.How{
					strategy.Limit(2),
				},
			},
			want: errToExpect,
		},
		{
			name: "not ready with deadline",
			fields: fields{
				checkers: []*state{
					{
						checker: &neverReady{err: errToExpect},
						name:    "never-ready",
					},
				},
				strategies: retry.How{
					strategy.Wait(90 * time.Millisecond),
				},
			},
			want: context.DeadlineExceeded,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Readiness{
				checkers:   tt.fields.checkers,
				strategies: tt.fields.strategies,
			}
			timeout, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			logger.SetNewGlobalLoggerQuietly(logger.DefaultConfig())
			defer cancel()
			got := r.IsReady(timeout, Logger(logger.FromContext(timeout)))
			if tt.want == nil {
				require.NoError(t, got)
			} else {
				require.Error(t, got)
				require.ErrorIs(t, got, tt.want)
			}
		})
	}
}
