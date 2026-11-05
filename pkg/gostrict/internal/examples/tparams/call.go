// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tparams

func callArea[S shape](s S) int32 {
	return s.area()
}

// Call calls the method of a type parameter directly.
func Call(n int32) int32 {
	return callArea(sq{n: n})
}
