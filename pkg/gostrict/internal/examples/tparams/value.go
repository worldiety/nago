// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tparams

func valueArea[S shape](s S) int32 {
	f := s.area
	return f()
}

// Value calls the method value of a type parameter.
func Value(n int32) int32 {
	return valueArea(sq{n: n})
}
