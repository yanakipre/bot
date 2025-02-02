package dynamicratelimiter

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/yanakipe/bot/internal/encodingtooling"
	"github.com/yanakipe/bot/internal/resttooling/ratelimiter"
	"github.com/yanakipe/bot/internal/testtooling"
)

func TestRateLimiter(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	rlm, err := NewRateLimitManager(ctx, func() (ratelimiter.RateLimitConfig, error) {
		return ratelimiter.RateLimitConfig{
			// 40 requests per 2 seconds means rate = 1 request per 50ms without burst
			Requests: 40,
			Period:   encodingtooling.Duration{Duration: 2 * time.Second},
			Burst:    20,
		}, nil
	}, 30*time.Second)
	require.NoError(t, err)

	responses := map[bool]int{}
	for i := 0; i < 50; i++ {
		responses[rlm.Allow("test")]++
	}
	require.Equal(t, 20, responses[true])
	require.Equal(t, 30, responses[false])

	// >= 10 tokens should be available after 500ms as rate limit is 20 requests per second
	time.Sleep(500 * time.Millisecond)

	responses = map[bool]int{}
	for i := 0; i < 50; i++ {
		responses[rlm.Allow("test")]++
	}
	require.GreaterOrEqual(t, 10, responses[true])
}

func TestRateLimiterConfigUpdate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testtooling.SetNewGlobalLoggerQuietly()

	iteration := 0
	rlm, err := NewRateLimitManager(ctx, func() (ratelimiter.RateLimitConfig, error) {
		if iteration == 0 {
			iteration++
			return ratelimiter.RateLimitConfig{
				Requests: 20,
				Period:   encodingtooling.Duration{Duration: 2 * time.Second},
				Burst:    20,
			}, nil
		} else {
			return ratelimiter.RateLimitConfig{
				Requests: 10,
				Period:   encodingtooling.Duration{Duration: 1 * time.Second},
				Burst:    10,
			}, nil
		}
	},
		// Refresh config every 1 second
		1*time.Second,
	)
	require.NoError(t, err)

	responses := map[bool]int{}
	for i := 0; i < 50; i++ {
		responses[rlm.Allow("test")]++
	}
	require.Equal(t, 20, responses[true])
	require.Equal(t, 30, responses[false])

	time.Sleep(1500 * time.Millisecond)

	responses = map[bool]int{}
	for i := 0; i < 50; i++ {
		responses[rlm.Allow("test")]++
	}
	require.Equal(t, 10, responses[true])
	require.Equal(t, 40, responses[false])
}
