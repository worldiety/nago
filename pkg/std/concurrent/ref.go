// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package concurrent

import (
	"sync/atomic"
)

// Value is just a box which updates it value atomically.
type Value[T any] struct {
	v atomic.Pointer[T]
}

func (v *Value[T]) Value() T {
	x := v.v.Load()
	if x == nil {
		var zero T
		return zero
	}

	return *x
}

func (v *Value[T]) SetValue(val T) {
	v.v.Store(&val)
}
