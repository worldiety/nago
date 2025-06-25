// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package async

import (
	"go.wdy.de/nago/pkg/std/concurrent"
	"sync"
)

// A Future represents a result which may occur in the future or not at all. It does intentionally not
// provide any blocking waits or timeout constructs, because that often is not what you actually want.
// Especially blocking in any event loops will easily cause deadlocks of all kinds.
type Future[T any] struct {
	mutex   sync.Mutex
	value   T
	err     error
	done    bool
	m       concurrent.RWMap[int64, func(T, error)]
	nextObs int64
}

func (f *Future[T]) Observe(fn func(t T, err error)) (close func()) {
	f.mutex.Lock()
	val := f.value
	err := f.err
	done := f.done
	ptr := f.nextObs + 1
	if !done {
		f.m.Put(ptr, fn)
	}
	f.mutex.Unlock()

	if done {
		fn(val, err)
		return func() {}
	}

	return func() {
		f.m.Delete(ptr)
	}
}

func (f *Future[T]) Set(value T, err error) {
	f.mutex.Lock()
	f.value = value
	f.err = err
	f.done = true
	f.mutex.Unlock()

	for _, fn := range f.m.All() {
		fn(value, err)
	}
}

func (f *Future[T]) Done() bool {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	return f.done
}

func (f *Future[T]) Get() (T, bool) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.done && f.err == nil {
		return f.value, true
	}

	var zero T
	return zero, false
}

func (f *Future[T]) Err() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.err
}
