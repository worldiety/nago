// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package crud

// Property is the contract for model conversions between the crud UI components and individual domain models.
// See also [Ptr] for a simple pointer-based adapter, if nothing else than a simple pointer is required for
// model transformation. See also [PropertyFuncs] for a quick inline-adapter implementation using closures.
type Property[E any, T any] interface {
	Set(dst *E, v T)
	Get(*E) T
}

// Ptr returns a get/set Property implementation based on simple (field) pointer from a (domain) model (struct).
func Ptr[E, T any](f func(model *E) *T) Property[E, T] {
	return fieldPtr[E, T](f)
}

type fieldPtr[E any, T any] func(model *E) *T

func (f fieldPtr[E, T]) Set(dst *E, v T) {
	fieldPtr := f(dst)
	*fieldPtr = v
}

func (f fieldPtr[E, T]) Get(e *E) T {
	fieldPtr := f(e)
	return *fieldPtr
}

type fnHolder[E any, T any] struct {
	get func(*E) T
	set func(dst *E, v T)
}

func (f fnHolder[E, T]) Set(dst *E, v T) {
	f.set(dst, v)
}

func (f fnHolder[E, T]) Get(e *E) T {
	return f.get(e)
}

// PropertyFuncs wraps a getter and setter into a property. The get and set funcs can be nil. If the getter is nil,
// the zero value of T is returned. If setter is nil, it does nothing.
func PropertyFuncs[E any, T any](get func(*E) T, set func(dst *E, v T)) Property[E, T] {
	if get == nil {
		get = func(e *E) T {
			var zero T
			return zero
		}
	}

	if set == nil {
		set = func(e *E, v T) {
			// no-op
		}
	}

	return fnHolder[E, T]{
		get: get,
		set: set,
	}
}
