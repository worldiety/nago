// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tparams

type holder[S shape] struct{ s S }

func (h holder[S]) area() int32 {
	f := h.s.area
	return f()
}

// Field calls the method value of a field, whose type is a type parameter of a generic type.
func Field(n int32) int32 {
	return holder[sq]{s: sq{n: n}}.area()
}
