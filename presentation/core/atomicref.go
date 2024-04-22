package core

import "sync"

type AtomicRef[T any] struct {
	mutex sync.Mutex
	v     T
}

func NewAtomicRef[T any]() *AtomicRef[T] {
	return &AtomicRef[T]{}
}

func (a *AtomicRef[T]) Load() T {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.v
}

func (a *AtomicRef[T]) Unload() T {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	tmp := a.v

	var zero T
	a.v = zero

	return tmp
}

func (a *AtomicRef[T]) Store(v T) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.v = v
}

func (a *AtomicRef[T]) With(fn func(T) T) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.v = fn(a.v)
}
