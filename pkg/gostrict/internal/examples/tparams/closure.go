// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tparams

func closureArea[S shape](s S) int32 {
	f := s.area
	g := func() int32 { return 2 * f() }
	return g()
}

// Closure calls the method value of a type parameter in a function literal.
func Closure(n int32) int32 {
	return closureArea(sq{n: n})
}
