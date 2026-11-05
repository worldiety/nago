// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package stdvalues

import "go.wdy.de/nago/pkg/gostrict/std/list"

// SortInfer sorts through list.SortFunc, instantiated by the type of the variable, which it is assigned to.
func SortInfer(a, b int32) int32 {
	l := list.Of(a, b, a+b)
	var s func(list.List[int32], func(x, y int32) int32) = list.SortFunc
	s(l, func(x, y int32) int32 { return y - x })
	v, _ := l.Get(0)
	w, _ := l.Get(2)
	return 10*v + w
}
