package xmaps

import (
	"iter"
	"sync"
)

type ConcurrentMap[K, V any] struct {
	m *sync.Map
}

func NewConcurrentMap[K, V any]() *ConcurrentMap[K, V] {
	return &ConcurrentMap[K, V]{
		m: new(sync.Map),
	}
}

func (m *ConcurrentMap[K, V]) Store(key K, value V) {
	m.m.Store(key, value)
}

func (m *ConcurrentMap[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	v, ok := m.m.LoadOrStore(key, value)
	return v.(V), ok
}

func (m *ConcurrentMap[K, V]) Load(key K) (value V, ok bool) {
	v, ok := m.m.Load(key)
	if !ok {
		var zero V
		return zero, false
	}

	return v.(V), ok
}

func (m *ConcurrentMap[K, V]) Delete(key K) {
	m.m.Delete(key)
}

// All does not block any other calls.
func (m *ConcurrentMap[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		m.m.Range(func(k, v interface{}) bool {
			return yield(k.(K), v.(V))
		})
	}

}

func (m *ConcurrentMap[K, V]) Clear() {
	m.m.Clear()
}
