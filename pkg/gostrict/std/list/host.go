// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package list

import "reflect"

// The methods of this file convert the lists for a host in Go, e.g. into the values of an engine, see
// gostrict.ToNeutral. The subset does not see them.

// HostValues returns the elements.
func (l List[T]) HostValues() []any {
	return hostValues(l.p)
}

// HostFromValues returns a list of the values, which are of type T.
func (List[T]) HostFromValues(values []any) any {
	return List[T]{p: hostFrom[T](values)}
}

// HostElem returns the type of the elements.
func (List[T]) HostElem() reflect.Type {
	return reflect.TypeFor[T]()
}

// HostValues returns the elements.
func (f Frozen[T]) HostValues() []any {
	return hostValues(f.p)
}

// HostFromValues returns a frozen list of the values, which are of type T.
func (Frozen[T]) HostFromValues(values []any) any {
	return Frozen[T]{p: hostFrom[T](values)}
}

// HostElem returns the type of the elements.
func (Frozen[T]) HostElem() reflect.Type {
	return reflect.TypeFor[T]()
}

func hostValues[T any](p *[]T) []any {
	if p == nil {
		return []any{}
	}
	res := make([]any, len(*p))
	for i, v := range *p {
		res[i] = v
	}
	return res
}

func hostFrom[T any](values []any) *[]T {
	s := make([]T, len(values))
	for i, v := range values {
		s[i] = v.(T)
	}
	return &s
}
