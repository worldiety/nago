// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tparams

func passApply(f func() int32) int32 {
	return f() + 1
}

func passArea[S shape](s S) int32 {
	return passApply(s.area)
}

// Pass passes the method value of a type parameter to a function.
func Pass(n int32) int32 {
	return passArea(sq{n: n})
}
