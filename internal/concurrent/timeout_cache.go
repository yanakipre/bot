package concurrent

import "time"

type cacheValue[V any] struct {
	val        V
	validUntil time.Time
}

// TimeoutCache is a cache that stores values for a limited time.
// If the value is not in the cache or has expired, it fetches the value using the fetcher function.
type TimeoutCache[K comparable, V any] struct {
	m          *Map[K, cacheValue[V]]
	now        func(K) time.Time
	validUntil func(K) time.Time
}

func NewTimeoutCache[K comparable, V any](
	now func(K) time.Time,
	validUntil func(K) time.Time,
) *TimeoutCache[K, V] {
	return &TimeoutCache[K, V]{
		m:          NewMap[K, cacheValue[V]](),
		validUntil: validUntil,
		now:        now,
	}
}

func (c *TimeoutCache[K, V]) Get(key K, fetcher func(K) (V, error)) (V, error) {
	value, ok := c.m.Get(key)
	if ok && c.now(key).Before(value.validUntil) {
		return value.val, nil
	}
	val, err := fetcher(key)
	if err != nil {
		return val, err
	}

	c.m.Add(key, cacheValue[V]{
		val:        val,
		validUntil: c.validUntil(key),
	})

	return val, nil
}
