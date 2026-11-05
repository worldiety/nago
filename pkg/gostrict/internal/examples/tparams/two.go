// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tparams

func twoArea[S shape](s S) int32 {
	f := s.area
	return f()
}

// Two calls the method value of a type parameter with two type arguments.
func Two(n int32) int32 {
	return twoArea(sq{n: n}) + twoArea(&rect{w: n, h: 2})
}
