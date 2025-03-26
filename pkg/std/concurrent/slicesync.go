// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package concurrent

import "sync"

type Slice[T any] struct {
	mutex sync.Mutex
	slice []T
}

// Len does not allocate. Note, that Len does not make much sense in concurrent situations.
func (l *Slice[T]) Len() int {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return len(l.slice)
}

// Append locks and appends the given set.
func (l *Slice[T]) Append(v ...T) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	tmp := make([]T, len(l.slice), len(l.slice)+len(v))
	copy(tmp, l.slice)
	tmp = append(tmp, v...)
	l.slice = tmp
}

func (l *Slice[T]) Clear() {
	l.PopAll()
}

// PopAll returns the underlying slice and moves it ownership to the caller. Afterwards, the internal slice is empty.
func (l *Slice[T]) PopAll() []T {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	tmp := l.slice
	l.slice = nil
	return tmp
}
