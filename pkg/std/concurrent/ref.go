package concurrent

import "sync"

// Value is just a box which updates it value atomically.
type Value[T any] struct {
	mutex sync.Mutex
	v     T
}

func (v *Value[T]) Value() T {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	return v.v
}

func (v *Value[T]) SetValue(val T) {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	v.v = val
}

func CompareAndSwap[T comparable](v *Value[T], old, new T) (swapped bool) {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	if v.v != old {
		return false
	}

	v.v = new
	return true
}
