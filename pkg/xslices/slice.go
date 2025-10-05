// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package xslices

import (
	"encoding/json"
	"iter"
	"slices"
)

// Slice is an immutable view into a conventional slice. Note, that only variables used for unmarshalling
// will implicitly use a pointer to change its own value.
type Slice[T any] struct {
	s []T
}

// Wrap takes ownership of the given slice without an additional copy. See also [New]. This is inherently unsafe.
func Wrap[T any](v ...T) Slice[T] {
	return Slice[T]{v}
}

// New creates a copy of the given slice. See also [Wrap].
func New[T any](s ...T) Slice[T] {
	return Slice[T]{slices.Clone(s)}
}

func (s Slice[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.s)
}

func (s *Slice[T]) UnmarshalJSON(data []byte) error {
	s.s = nil // intentionally discard that reference to ensure that shared slices will not get accidentally corrupted
	return json.Unmarshal(data, &s.s)
}

func (s Slice[T]) At(idx int) T {
	return s.s[idx]
}

func (s Slice[T]) Len() int {
	return len(s.s)
}

func (s Slice[T]) All() iter.Seq[T] {
	return slices.Values(s.s)
}

// Clone returns a shallow clone of the underlying slice. If the slice values contains any pointer types, the mutation
// will still be aliased.
func (s Slice[T]) Clone() []T {
	return slices.Clone(s.s)
}

// Append copies all values into a new underlying slice and returns a new immutable slice.
func (s Slice[T]) Append(v ...T) Slice[T] {
	if len(s.s) == 0 {
		return Slice[T]{slices.Clone(v)}
	}

	// if we don't do that, the underlying spare space may get aliased and reused between different slices
	alloc := make([]T, len(s.s)+len(v))
	copy(alloc, s.s)
	copy(alloc[len(s.s):], v)
	return Slice[T]{alloc}
}

func (s Slice[T]) IsZero() bool {
	return len(s.s) == 0
}

func Contains[T comparable](s Slice[T], v T) bool {
	for _, i := range s.s {
		if i == v {
			return true
		}
	}

	return false
}
