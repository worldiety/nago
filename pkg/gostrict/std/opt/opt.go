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
// OrZero and OrZeroErr take the results of a call, which returns (T, bool) or (T, error), and return the value or
// the zero value, e.g. opt.OrZero(l.Get(i)) or opt.OrZeroErr(text.ParseInt(s)), where the zero value is the wanted
// default. The subset allows them only for a T, whose zero value contains no nil; for any other T use GetOr of a
// list or a dict with an explicit default or check ok: if v, ok := l.Get(i); ok { ... }.
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

// OrZero returns v, if ok is true, or the zero value of T, e.g. opt.OrZero(l.Get(i)) or opt.OrZero(m.Get(k)). The
// subset rejects it for a T, whose zero value contains nil, e.g. a list, a map, a pointer, an interface or a struct
// with such a field: use l.GetOr(i, def) or if v, ok := l.Get(i); ok { ... } for these.
func OrZero[T any](v T, ok bool) T {
	if !ok {
		var zero T
		return zero
	}
	return v
}

// OrZeroErr returns v, if err is nil, or the zero value of T, e.g. opt.OrZeroErr(text.ParseInt(s)). The subset
// rejects it for a T, whose zero value contains nil, like [OrZero]: check the error for these, if v, err := f();
// err == nil { ... }.
func OrZeroErr[T any](v T, err error) T {
	if err != nil {
		var zero T
		return zero
	}
	return v
}
