package rate

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/yanakipre/bot/internal/encodingtooling"
)

// These tests could fail due to leap seconds - it's unlikely, but let's use a frozen time to get around that.
var now, _ = time.Parse(time.RFC3339, "2024-07-18T12:00:00Z")

func Test_window_allow(t *testing.T) {
	tests := map[string]struct {
		window       window
		wantAllowed  bool
		wantDuration time.Duration
	}{
		"allows requests under the limit": {
			window: window{
				limit:    10,
				duration: time.Second,
				start:    now,
				requests: 9,
			},
			wantAllowed:  true,
			wantDuration: 0,
		},
		"allows requests if the window has expired": {
			window: window{
				limit:    10,
				duration: time.Second,
				start:    now.Add(-time.Second),
				requests: 10,
			},
			wantAllowed:  true,
			wantDuration: 0,
		},
		"denies requests over the limit": {
			window: window{
				limit:    10,
				duration: time.Second,
				start:    now,
				requests: 10,
			},
			wantAllowed:  false,
			wantDuration: time.Second,
		},
		"returns the time left for the window to expire": {
			window: window{
				limit:    10,
				duration: time.Hour,
				// An unusual number, on purpose - to see if we return the complement of it to form a full hour.
				start:    now.Add(-27 * time.Minute),
				requests: 10,
			},
			wantAllowed:  false,
			wantDuration: 33 * time.Minute,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotAllowed, gotDuration := tt.window.allow(now)

			require.Equal(t, tt.wantAllowed, gotAllowed)
			require.Equal(t, tt.wantDuration, gotDuration)
		})
	}
}

func Test_windows_allow(t *testing.T) {
	tests := map[string]struct {
		windows      windows
		wantAllowed  bool
		wantDuration time.Duration
	}{
		"allows requests under the limit of all windows": {
			windows: windows{{
				limit:    10,
				duration: time.Second,
				requests: 9,
			}, {
				limit:    300,
				duration: time.Minute,
				requests: 299,
			}, {
				limit:    1500,
				duration: 10 * time.Minute,
				requests: 1499,
			}},
			wantAllowed:  true,
			wantDuration: 0,
		},
		"denies requests if the first limit is reached": {
			windows: windows{{
				limit:    10,
				duration: time.Second,
				requests: 10,
			}, {
				limit:    300,
				duration: time.Minute,
				requests: 299,
			}, {
				limit:    1500,
				duration: 10 * time.Minute,
				requests: 1499,
			}},
			wantAllowed:  false,
			wantDuration: time.Second,
		},
		"denies requests if the second limit is reached": {
			windows: windows{{
				limit:    10,
				duration: time.Second,
				requests: 9,
			}, {
				limit:    300,
				duration: time.Minute,
				requests: 300,
			}, {
				limit:    1500,
				duration: 10 * time.Minute,
				requests: 1499,
			}},
			wantAllowed:  false,
			wantDuration: time.Minute,
		},
		"denies requests if the third limit is reached": {
			windows: windows{{
				limit:    10,
				duration: time.Second,
				requests: 9,
			}, {
				limit:    300,
				duration: time.Minute,
				requests: 299,
			}, {
				limit:    1500,
				duration: 10 * time.Minute,
				requests: 1500,
			}},
			wantAllowed:  false,
			wantDuration: 10 * time.Minute,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			for _, window := range tt.windows {
				window.start = now
			}

			gotAllowed, gotDuration := tt.windows.allow(now)

			require.Equal(t, tt.wantAllowed, gotAllowed)
			require.Equal(t, tt.wantDuration, gotDuration)
		})
	}
}

func Test_MultiBucketFixedWindowLimiter_Overrides(t *testing.T) {
	now := time.Now()
	mb := NewMultiBucketFixedWindowLimiter([]WindowConfig{
		{Limit: 10, Duration: encodingtooling.NewDuration(time.Second)},
	})

	key := "some-rate-limiting-key"

	// We'll start by allowing 10 requests.
	for range 10 {
		allowed, wait := mb.Allow(key, now)
		require.True(t, allowed)
		require.Equal(t, time.Duration(0), wait)
	}

	// The 11th request should be denied.
	allowed, wait := mb.Allow(key, now)
	require.False(t, allowed)
	require.Equal(t, time.Second, wait)

	// Now we'll add an override for this key.
	mb.OverrideWindows(key, []WindowConfig{
		{Limit: 20, Duration: encodingtooling.NewDuration(time.Second)},
	})

	// Then we'll see if WouldAllow would allow the 11th request.
	allowed, wait = mb.WouldAllow(key, now)
	require.True(t, allowed)
	require.Equal(t, time.Duration(0), wait)

	// And the next 10 requests should be allowed.
	for range 10 {
		allowed, wait := mb.Allow(key, now)
		require.True(t, allowed)
		require.Equal(t, time.Duration(0), wait)
	}

	// But the 22nd request should be denied.
	allowed, wait = mb.Allow(key, now)
	require.False(t, allowed)
	require.Equal(t, time.Second, wait)
}
