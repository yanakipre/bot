package concurrent

import (
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeoutCache_Get(t *testing.T) {
	// Create a TimeoutCache instance
	now := lo.ToPtr(time.Now())
	cache := NewTimeoutCache[string, string](
		func(_ string) time.Time { return *now },
		func(_ string) time.Time { return now.Add(time.Minute) },
	)

	// Test case 1: Value doesn't exist in the cache
	result, err := cache.Get("key1", func(s string) (string, error) {
		return "val1", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "val1", result)
	require.Equal(t, now.Add(time.Minute), cache.validUntil("key1"))

	// Test case 2: Value exists in the cache and has not expired
	// Mock 10 seconds passing
	*now = now.Add(10 * time.Second)
	result, err = cache.Get("key1", func(s string) (string, error) {
		return "val2", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "val1", result)

	// Test case 3: Value has expired
	// Mock passing 100 seconds
	*now = now.Add(100 * time.Second)
	result, err = cache.Get("key1", func(s string) (string, error) {
		return "val3", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "val3", result)
}
