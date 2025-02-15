package pgretries

import (
	"context"
	"time"

	"github.com/kamilsk/retry/v5"
	"github.com/kamilsk/retry/v5/backoff"
	"github.com/kamilsk/retry/v5/strategy"
	"go.uber.org/zap"

	"github.com/yanakipre/bot/internal/logger"
	"github.com/yanakipre/bot/internal/retrytooling"
)

type Retry struct {
	how retry.How
}

func (r *Retry) Do(ctx context.Context, f func(ctx context.Context) error) error {
	return retry.Do(ctx, f, r.how...)
}

// Strategy for retrying errors
func Strategy() retry.How {
	return retry.How{
		func(br retry.Breaker, attempt uint, err error) bool {
			if attempt == 0 {
				return true
			}
			ctx, ok := br.(context.Context)
			if !ok {
				panic("context is not passed to retry.Breaker, should never happen")
			}
			retryable := isRetryable(err)
			if retryable {
				logger.FromContext(ctx).Info("rdb error is retryable", zap.Error(err), zap.Uint("attempt", attempt))
			}
			return retryable
		},
		strategy.BackoffWithJitter(
			backoff.Fibonacci(10*time.Millisecond),
			retrytooling.NormalDistribution(0.25),
		),
	}
}

func NoOp() retry.How {
	return retry.How{strategy.Limit(1)}
}
