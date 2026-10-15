// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package opt provides an optional value of the gostrict subset. Its zero value is empty, so a struct may refer to
// itself, which a pointer, that must never be nil, does not allow:
//
//	type Node struct {
//		Value int32
//		Next  opt.Option[*Node]
//	}
//
// The value is read only with Get, whose result is valid only where ok is true, or with Or. An option can not
// change, so it is shared freely, and it can not be compared.
//
//gostrict:header
package opt

// Option holds a value or nothing.
type Option[T any] struct {
	_  [0]func()
	v  T
	ok bool
}

// Some returns an option with the value.
func Some[T any](v T) Option[T] {
	return Option[T]{v: v, ok: true}
}

// None returns an empty option, which is also the zero value.
func None[T any]() Option[T] {
	return Option[T]{}
}

// Get returns the value and true or the zero value and false, if the option is empty.
func (o Option[T]) Get() (T, bool) {
	return o.v, o.ok
}

// Or returns the value or def, if the option is empty.
func (o Option[T]) Or(def T) T {
	if !o.ok {
		return def
	}
	return o.v
}

// IsSome tells, whether the option holds a value.
func (o Option[T]) IsSome() bool {
	return o.ok
}
