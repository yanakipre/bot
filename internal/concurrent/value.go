package concurrent

import (
	"sync"
)

type Value[V any] struct {
	mu    sync.RWMutex
	value V
}

func NewValue[V any](value V) *Value[V] {
	return &Value[V]{
		value: value,
	}
}

func (v *Value[V]) Get() V {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.value
}

func (v *Value[V]) Update(f func(V) V) {
	v.mu.Lock()
	v.value = f(v.value)
	v.mu.Unlock()
}
