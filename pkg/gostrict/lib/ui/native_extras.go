// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package ui

import (
	"go.wdy.de/nago/application/image"
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/presentation/core"
)

func (s IndicesState) nago() *core.State[[]int] {
	return get[*core.State[[]int]](s.ref)
}

func indicesOf(v []int) list.List[int32] {
	res := list.New[int32]()
	for _, i := range v {
		res.Append(int32(i))
	}
	return res
}

func intsOf(l list.List[int32]) []int {
	var res []int
	for i := range l.Values() {
		res = append(res, int(i))
	}
	return res
}

func rowsUpTo(rows int32) []int {
	res := make([]int, max(rows, 0))
	for i := range res {
		res[i] = i
	}
	return res
}

func imageIDOf(id ImageID) image.ID {
	return image.ID(id)
}
