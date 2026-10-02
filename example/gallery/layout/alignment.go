// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"go.wdy.de/nago/example/gallery/app"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
)

func init() {
	app.Register("alignment", func(wnd core.Window) core.View {
		alignments := []Alignment{
			TopLeading, Top, TopTrailing,
			Leading, Center, Trailing,
			BottomLeading, Bottom, BottomTrailing,
		}

		return Grid(
			ForEach(alignments, func(a Alignment) TGridCell {
				return GridCell(
					VStack(Text(a.String())).
						Alignment(a).
						BackgroundColor("#C9E7F8").
						Padding(Padding{}.All(L8)).
						Frame(Frame{}.Size(L200, L96)),
				)
			})...,
		).Columns(3).Gap(L8)
	})
}
