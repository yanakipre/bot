package concurrent

import (
	"fmt"
	"sync"

	"github.com/samber/lo"
)

type Map[K comparable, V any] struct {
	m     sync.RWMutex
	cache map[K]V
}

func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{
		cache: make(map[K]V),
	}
}

func (c *Map[K, V]) Get(key K) (V, bool) {
	c.m.RLock()
	v, ok := c.cache[key]
	c.m.RUnlock()
	return v, ok
}

func (c *Map[K, V]) Add(key K, value V) {
	c.m.Lock()
	c.cache[key] = value
	c.m.Unlock()
}

func (c *Map[K, V]) Set(m map[K]V) {
	c.m.Lock()
	c.cache = m
	c.m.Unlock()
}

func (c *Map[K, V]) InitBulk(keys []K, values []V) {
	if len(keys) != len(values) {
		panic(fmt.Sprintf("keys and values length mismatch: %d != %d", len(keys), len(values)))
	}
	c.m.Lock()
	c.cache = make(map[K]V, len(keys))
	for i := range keys {
		c.cache[keys[i]] = values[i]
	}
	c.m.Unlock()
}

func (c *Map[K, V]) Remove(key K) {
	c.m.Lock()
	delete(c.cache, key)
	c.m.Unlock()
}

func (c *Map[K, V]) Len() int {
	c.m.RLock()
	l := len(c.cache)
	c.m.RUnlock()
	return l
}

func (c *Map[K, V]) GetAll() []V {
	c.m.RLock()
	vv := lo.MapToSlice(c.cache, func(_ K, v V) V {
		return v
	})
	c.m.RUnlock()
	return vv
}
