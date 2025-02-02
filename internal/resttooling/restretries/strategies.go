package restretries

import (
	"math/rand"
	"time"

	"github.com/kamilsk/retry/v5"
	"github.com/kamilsk/retry/v5/backoff"
	"github.com/kamilsk/retry/v5/jitter"
	"github.com/kamilsk/retry/v5/strategy"
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
			jitter.NormalDistribution(
				rand.New(rand.NewSource(time.Now().UnixNano())), //nolint:gosec
				0.25,
			),
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
