package retrytooling

import (
	"math"
	"math/rand/v2"
	"time"

	"github.com/kamilsk/retry/v5/jitter"
)

// Code below is copied from
// github.com/kamilsk/retry/v5/jitter
// The functions there expected a *rand.Rand, however the source for *rand.Rand
// is not goroutine safe, so we use the global rand instead.

// Deviation creates a Transformation that transforms a duration into a result
// duration that deviates from the input randomly by a given factor.
//
// The given generator is what is used to determine the random transformation.
//
// Inspired by https://developers.google.com/api-client-library/java/google-http-java-client/backoff
func Deviation(factor float64) jitter.Transformation {
	return func(duration time.Duration) time.Duration {
		min := int64(math.Floor(float64(duration) * (1 - factor)))
		max := int64(math.Ceil(float64(duration) * (1 + factor)))
		return time.Duration(rand.Int64N(max-min) + min) //nolint:gosec
	}
}

// NormalDistribution creates a Transformation that transforms a duration into a
// result duration based on a normal distribution of the input and the given
// standard deviation.
//
// The given generator is what is used to determine the random transformation.
func NormalDistribution(standardDeviation float64) jitter.Transformation {
	return func(duration time.Duration) time.Duration {
		return time.Duration(rand.NormFloat64()*standardDeviation + float64(duration)) //nolint:gosec
	}
}
