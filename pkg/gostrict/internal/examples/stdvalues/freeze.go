// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package stdvalues

import "go.wdy.de/nago/pkg/gostrict/std/list"

// Freeze freezes a list through list.Freeze as a value.
func Freeze(a, b int32) int32 {
	f := list.Freeze[int32]
	l := f(list.Of(a, b))
	v := l.GetOr(1, 0)
	return 10*l.Len() + v
}
