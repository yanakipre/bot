package restretries

import (
	"time"

	"github.com/kamilsk/retry/v5"
	"github.com/kamilsk/retry/v5/backoff"
	"github.com/kamilsk/retry/v5/strategy"

	"github.com/yanakipre/bot/internal/retrytooling"
)

type RetryStrategy retry.How

// FibonacciBackoffWithJitterRetryStrategy
// - backoff is based on fibonacci sequence
// - jitter has normal distribution
func FibonacciBackoffWithJitterRetryStrategy(
	initialBackoff time.Duration,
	attempts uint,
) RetryStrategy {
	return retry.How{
		strategy.Limit(attempts),
		strategy.BackoffWithJitter(
			backoff.Fibonacci(initialBackoff),
			retrytooling.NormalDistribution(0.25),
		),
	}
}

// FixedIntervalRetryStrategy defines indefinite retries using given interval.
// Can be used when you need time guarantees, just use it with context.WithTimeout().
func FixedIntervalRetryStrategy(attempts uint, interval time.Duration) RetryStrategy {
	return retry.How{
		strategy.Limit(attempts),
		strategy.Wait(interval),
	}
}

func mustNotRetryTerminalErrorsStrategy(_ retry.Breaker, _ uint, err error) bool {
	// return false for all permanent (non-retriable errors), and
	// false otherwise.
	return !IsPermanentError(err)
}
