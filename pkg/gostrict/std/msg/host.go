// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package msg

import "reflect"

// The methods of this file convert the outbox for a host in Go, see gostrict.ToNeutral. The subset does not see
// them.

// HostValues returns the messages.
func (o Out[T]) HostValues() []any {
	if o.p == nil {
		return []any{}
	}
	res := make([]any, len(*o.p))
	for i, v := range *o.p {
		res[i] = v
	}
	return res
}

// HostFromValues returns an outbox of the messages, which are of type T.
func (Out[T]) HostFromValues(values []any) any {
	s := make([]T, len(values))
	for i, v := range values {
		s[i] = v.(T)
	}
	return Out[T]{p: &s}
}

// HostElem returns the type of the messages.
func (Out[T]) HostElem() reflect.Type {
	return reflect.TypeFor[T]()
}
