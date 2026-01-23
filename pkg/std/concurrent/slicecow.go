// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package concurrent

import (
	"iter"
	"sync"
)

// CoWSlice is a copy-on-write slice.
type CoWSlice[T any] struct {
	mutex sync.RWMutex
	slice []T
}

// Len does not allocate. Note, that Len does not make much sense in concurrent situations.
func (l *CoWSlice[T]) Len() int {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	return len(l.slice)
}

// Append locks and copies the entire set. This is very expensive.
func (l *CoWSlice[T]) Append(v ...T) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	tmp := make([]T, len(l.slice), len(l.slice)+len(v))
	copy(tmp, l.slice)
	tmp = append(tmp, v...)
	l.slice = tmp
}

func (l *CoWSlice[T]) InsertFirst(v T) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.slice = append([]T{v}, l.slice...)
}

// InsertAfterFunc locks and inserts the returned value after each value that returns true from the given function.
func (l *CoWSlice[T]) InsertAfterFunc(fn func(T) (T, bool)) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	tmp := make([]T, 0, len(l.slice)+1)
	for _, t := range l.slice {
		tmp = append(tmp, t)
		if v, ok := fn(t); ok {
			tmp = append(tmp, v)
		}

	}

	l.slice = tmp
}

func (l *CoWSlice[T]) Clear() {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.slice = nil
}

// Deprecated: use All
func (l *CoWSlice[T]) Each(yield func(T) bool) {
	l.mutex.RLock()
	ref := l.slice
	l.mutex.RUnlock()

	for _, t := range ref {
		if !yield(t) {
			return
		}
	}
}

// All iterates over all items. This is cheap and does not allocate and can never deadlock. Mutators will
// allocate new slices underneath, so any read always iterates on an immutable copy.
func (l *CoWSlice[T]) All() iter.Seq[T] {
	l.mutex.RLock()
	ref := l.slice
	l.mutex.RUnlock()

	return func(yield func(T) bool) {
		for _, t := range ref {
			if !yield(t) {
				return
			}
		}
	}
}
