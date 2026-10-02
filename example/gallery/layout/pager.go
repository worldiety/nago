// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"fmt"

	"go.wdy.de/nago/example/gallery/app"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/pager"
)

func init() {
	app.Register("pager", func(wnd core.Window) core.View {
		pageIdx := core.AutoState[int](wnd).Init(func() int {
			return 2
		})

		return VStack(
			Text(fmt.Sprintf("You are on page index %d.", pageIdx.Get())),
			pager.Pager(pageIdx).Count(5),
		).Gap(L8)
	})
}
