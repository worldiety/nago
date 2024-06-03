package concurrent

import (
	"go.wdy.de/nago/pkg/iter"
	"maps"
	"sync"
)

// CoWMap is a copy-on-write map.
type CoWMap[K comparable, V any] struct {
	m     map[K]V
	mutex sync.Mutex
}

// Put is most expensive, because it copies the entire internal map.
func (c *CoWMap[K, V]) Put(key K, value V) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	tmp := maps.Clone(c.m)
	if tmp == nil {
		tmp = map[K]V{}
	}
	tmp[key] = value
	c.m = tmp
}

func (c *CoWMap[K, V]) PutAll(m map[K]V) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	tmp := maps.Clone(c.m)
	if tmp == nil {
		tmp = map[K]V{}
	}
	for k, v := range m {
		tmp[k] = v
	}

	c.m = tmp
}

func (c *CoWMap[K, V]) PutIter(it iter.Seq2[K, V]) {
	// we would deadlock if applied on ourself or block infinite if the iter sleeps
	outSideLock := map[K]V{}
	it(func(k K, v V) bool {
		outSideLock[k] = v
		return true
	})

	c.mutex.Lock()
	defer c.mutex.Unlock()

	tmp := maps.Clone(c.m)
	if tmp == nil {
		tmp = map[K]V{}
	}
	for k, v := range outSideLock {
		tmp[k] = v
	}

	c.m = tmp
}

func (c *CoWMap[K, V]) Get(key K) (value V, ok bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.m == nil {
		var zero V
		return zero, false
	}

	value, ok = c.m[key]
	return value, ok
}

func (c *CoWMap[K, V]) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.m = nil
}

// Delete is most expensive, because it copies the entire internal map.
func (c *CoWMap[K, V]) Delete(key K) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	tmp := maps.Clone(c.m)
	if tmp == nil {
		return
	}
	delete(tmp, key)
	c.m = tmp
}

func (c *CoWMap[K, V]) Len() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.m == nil {
		return 0
	}

	return len(c.m)
}

// Each loops safely over a copy of the dataset without ever blocking or deadlocking.
func (c *CoWMap[K, V]) Each(yield func(key K, value V) bool) {
	c.mutex.Lock()
	ref := c.m
	c.mutex.Unlock()

	if ref == nil {
		return
	}

	for k, v := range ref {
		if !yield(k, v) {
			return
		}
	}
}
