// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package stdvalues

import "go.wdy.de/nago/pkg/gostrict/std/list"

// SortFunc sorts through list.SortFunc, explicitly instantiated as a value.
func SortFunc(a, b int32) int32 {
	l := list.Of(a, b, a+b)
	s := list.SortFunc[int32]
	s(l, func(x, y int32) int32 { return y - x })
	v := l.GetOr(0, 0)
	w := l.GetOr(2, 0)
	return 10*v + w
}
