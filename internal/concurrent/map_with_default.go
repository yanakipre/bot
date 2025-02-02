package concurrent

import (
	"sync"
)

type MapWithConstructor[K comparable, V any] struct {
	m           sync.RWMutex
	constructor func() V
	cache       map[K]V
}

func NewMapWithConstructor[K comparable, V any](f func() V) *MapWithConstructor[K, V] {
	return &MapWithConstructor[K, V]{
		cache:       make(map[K]V),
		constructor: f,
	}
}

func (c *MapWithConstructor[K, V]) Get(key K) V {
	c.m.Lock()
	defer c.m.Unlock()
	v, ok := c.cache[key]
	if !ok {
		v = c.constructor()
		c.cache[key] = v
	}
	return v
}
