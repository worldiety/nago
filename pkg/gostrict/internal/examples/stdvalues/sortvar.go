// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package stdvalues

import "go.wdy.de/nago/pkg/gostrict/std/list"

// sorter is list.SortFunc as the value of a package variable.
var sorter = list.SortFunc[int32]

// SortVar sorts through list.SortFunc in a package variable.
func SortVar(a, b int32) int32 {
	l := list.Of(a, b, a+b)
	sorter(l, func(x, y int32) int32 { return y - x })
	v := l.GetOr(0, 0)
	w := l.GetOr(2, 0)
	return 10*v + w
}
