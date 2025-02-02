package concurrent

import "time"

// SingleTimeoutCache is a cache that stores a single value for a limited time.
// If the value is not in the cache or has expired, it fetches the value using the fetcher function.
// XXX: We are using a map with a single key of type struct{} to store the value. So, not the optimal
// way to store a single value but it is a simple way to implement a single value cache using existing TimeoutCache.
type SingleTimeoutCache[V any] struct {
	cache *TimeoutCache[struct{}, V]
}

func NewSingleTimeoutCache[V any](
	now func() time.Time,
	ttl time.Duration,
) *SingleTimeoutCache[V] {
	return &SingleTimeoutCache[V]{
		cache: NewTimeoutCache[struct{}, V](
			func(_ struct{}) time.Time {
				return now()
			},
			func(_ struct{}) time.Time {
				return now().Add(ttl)
			},
		),
	}
}

func (c *SingleTimeoutCache[V]) Get(fetcher func() (V, error)) (V, error) {
	return c.cache.Get(struct{}{}, func(_ struct{}) (V, error) {
		return fetcher()
	})
}
