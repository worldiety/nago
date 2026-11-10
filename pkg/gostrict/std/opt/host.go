// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package opt

import "reflect"

// The methods of this file create options for a host in Go, which only knows the type at runtime, e.g. package json.
// The subset does not see them.

// HostFromOption returns Some of the value, which is of type T, or None.
func (Option[T]) HostFromOption(v any, ok bool) any {
	if !ok {
		return Option[T]{}
	}
	return Some(v.(T))
}

// HostOption returns the value and whether it is present.
func (o Option[T]) HostOption() (any, bool) {
	return o.v, o.ok
}

// HostElem returns the type of the value.
func (Option[T]) HostElem() reflect.Type {
	return reflect.TypeFor[T]()
}
