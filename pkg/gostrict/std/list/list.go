// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package list provides the list of the gostrict subset, which replaces the slices of Go. A list has reference
// semantics like an array of JavaScript: a copy of a list refers to the same elements. The elements are copied
// in and out, so a struct, which is read with [List.Get], is a copy. No function panics: an index out of range
// is reported by the result.
//
//gostrict:header
package list

import (
	"iter"
	"slices"
)

// List is a growable sequence of elements. The subset requires a list to be created with [New], [Of] or [Make],
// so the zero value never occurs in a checked program.
type List[T any] struct {
	p *[]T
}

// New returns an empty list.
func New[T any]() List[T] {
	var s []T
	return List[T]{p: &s}
}

// Of returns a list of the given elements.
func Of[T any](items ...T) List[T] {
	s := slices.Clone(items)
	if s == nil {
		s = []T{}
	}
	return List[T]{p: &s}
}

// Make returns a list of n zero values, or an empty list, if n is negative. The subset allows it only for types,
// whose zero value contains no nil.
func Make[T any](n int32) List[T] {
	s := make([]T, max(n, 0))
	return List[T]{p: &s}
}

// Len returns the number of elements.
func (l List[T]) Len() int32 {
	if l.p == nil {
		return 0
	}
	return int32(len(*l.p))
}

func (l List[T]) valid(i int32) bool {
	return i >= 0 && i < l.Len()
}

// Get returns the element at index i and true, or the zero value and false, if i is out of range.
func (l List[T]) Get(i int32) (T, bool) {
	if !l.valid(i) {
		var zero T
		return zero, false
	}
	return (*l.p)[i], true
}

// Set replaces the element at index i and reports, whether i is in range.
func (l List[T]) Set(i int32, v T) bool {
	if !l.valid(i) {
		return false
	}
	(*l.p)[i] = v
	return true
}

// Append adds the elements at the end.
func (l List[T]) Append(items ...T) {
	if l.p == nil {
		return // unreachable in a checked program
	}
	*l.p = append(*l.p, items...)
}

// Pop removes and returns the last element and true, or the zero value and false, if the list is empty.
func (l List[T]) Pop() (T, bool) {
	if l.Len() == 0 {
		var zero T
		return zero, false
	}
	s := *l.p
	v := s[len(s)-1]
	*l.p = s[:len(s)-1]
	return v, true
}

// Delete removes the element at index i and reports, whether i is in range.
func (l List[T]) Delete(i int32) bool {
	if !l.valid(i) {
		return false
	}
	*l.p = slices.Delete(*l.p, int(i), int(i)+1)
	return true
}

// Clone returns a new list with copies of the elements.
func (l List[T]) Clone() List[T] {
	if l.p == nil {
		return List[T]{}
	}
	s := slices.Clone(*l.p)
	if s == nil {
		s = []T{}
	}
	return List[T]{p: &s}
}

// All iterates over the indices and elements, which the list has, when the iteration starts, like a range over a
// slice: elements appended during the iteration are not visited, so the iteration always ends.
func (l List[T]) All() iter.Seq2[int32, T] {
	return func(yield func(int32, T) bool) {
		n := l.Len()
		for i := int32(0); i < n && i < l.Len(); i++ {
			if !yield(i, (*l.p)[i]) {
				return
			}
		}
	}
}

// Values iterates over the elements, see [List.All].
func (l List[T]) Values() iter.Seq[T] {
	return func(yield func(T) bool) {
		n := l.Len()
		for i := int32(0); i < n && i < l.Len(); i++ {
			if !yield((*l.p)[i]) {
				return
			}
		}
	}
}

// SortFunc sorts the list stably by cmp, which returns a negative number, if a is less than b, a positive one,
// if it is greater and zero, if both are equal. Equal elements keep their order.
func SortFunc[T any](l List[T], cmp func(a, b T) int32) {
	if l.p == nil {
		return
	}
	slices.SortStableFunc(*l.p, func(a, b T) int { return int(cmp(a, b)) })
}

// Frozen is a list, which can not change, e.g. a table in a package variable, which the subset allows only for
// values, which can not change.
type Frozen[T any] struct {
	p *[]T
}

// Freeze returns a frozen copy of the list.
func Freeze[T any](l List[T]) Frozen[T] {
	c := l.Clone()
	if c.p == nil {
		var s []T
		return Frozen[T]{p: &s}
	}
	return Frozen[T]{p: c.p}
}

// Len returns the number of elements.
func (f Frozen[T]) Len() int32 {
	return List[T](f).Len()
}

// Get returns the element at index i and true, or the zero value and false, if i is out of range.
func (f Frozen[T]) Get(i int32) (T, bool) {
	return List[T](f).Get(i)
}

// All iterates over the indices and elements.
func (f Frozen[T]) All() iter.Seq2[int32, T] {
	return List[T](f).All()
}

// Values iterates over the elements.
func (f Frozen[T]) Values() iter.Seq[T] {
	return List[T](f).Values()
}
