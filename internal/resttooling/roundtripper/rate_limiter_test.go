package roundtripper

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

func TestRateLimiter(t *testing.T) {
	ctx := context.Background()

	// 40 requests per 2 seconds means rate = 1 request per 50ms without burst
	rl := rate.NewLimiter(rate.Every(2*time.Second/40), 1)
	wg := &sync.WaitGroup{}
	now := time.Now()

	// set 50 request to exclude flaps
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := rl.Wait(ctx)
			require.NoError(t, err)
		}()
	}
	wg.Wait()

	// 50 request should be done in ~ 2.5 seconds
	require.Greater(t, time.Since(now), 2*time.Second)
}
