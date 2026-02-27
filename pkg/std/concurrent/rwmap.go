package concurrent

import (
	"iter"
	"maps"
	"sync"
)

// RWMap is a simple rw mutex based map. This is probably the most memory efficient way and for mostly read
// workloads perhaps also the fastest choice. However, there is the danger of deadlocks for all
// callback based functions and you may want to consider [CoWMap].
type RWMap[K comparable, V any] struct {
	m     map[K]V
	mutex sync.RWMutex
}

func WrapRWMap[K comparable, V any](m map[K]V) *RWMap[K, V] {
	return &RWMap[K, V]{m: m}
}

func (c *RWMap[K, V]) Put(key K, value V) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.m == nil {
		c.m = make(map[K]V)
	}

	c.m[key] = value
}

// LoadOrStore either creates a new entry or returns an existing one.
func (c *RWMap[K, V]) LoadOrStore(key K, defaultV V) (actual V, loaded bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	v, ok := c.m[key]
	if ok {
		return v, true
	}

	if c.m == nil {
		c.m = make(map[K]V)
	}

	c.m[key] = defaultV
	return defaultV, false
}

func (c *RWMap[K, V]) Get(key K) (value V, ok bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.m == nil {
		var zero V
		return zero, false
	}

	v, ok := c.m[key]
	return v, ok
}

func (c *RWMap[K, V]) Delete(key K) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.m, key)
}

// DeleteFunc will hold the write lock until completion, thus do not call any map method, otherwise a deadlock
// occurs.
func (c *RWMap[K, V]) DeleteFunc(fn func(K, V) bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for k, v := range c.m {
		if fn(k, v) {
			delete(c.m, k)
		}
	}
}

// All iterates over all key-values within the map under the maps global read mutex. Note that you will
// cause a guaranteed deadlock if you do other reads or writes within the loop body.
func (c *RWMap[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		c.mutex.RLock()
		defer c.mutex.RUnlock()
		for k, v := range c.m {
			if !yield(k, v) {
				return
			}
		}
	}
}

func (c *RWMap[K, V]) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	clear(c.m)
}

func (c *RWMap[K, V]) Len() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.m)
}

func (c *RWMap[K, V]) Clone() *RWMap[K, V] {
	return &RWMap[K, V]{m: maps.Clone(c.m)}
}

func (c *RWMap[K, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		c.mutex.RLock()
		defer c.mutex.RUnlock()
		for _, v := range c.m {
			if !yield(v) {
				return
			}
		}
	}
}
