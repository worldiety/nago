// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tparams

func pointerArea[S shape](s S) int32 {
	f := s.area
	return f()
}

// Pointer calls the method value of a type parameter, whose type argument is a pointer with the method.
func Pointer(n int32) int32 {
	return pointerArea(&rect{w: n, h: n + 1})
}
