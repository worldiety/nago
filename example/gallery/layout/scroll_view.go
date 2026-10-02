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
)

func init() {
	app.Register("scroll-view", func(wnd core.Window) core.View {
		numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

		return ScrollView(
			HStack(
				ForEach(numbers, func(n int) core.View {
					return Text(fmt.Sprintf("Item %d", n)).
						BackgroundColor("#C9E7F8").
						Frame(Frame{}.Size(L120, L80))
				})...,
			).Gap(L8),
		).Axis(ScrollViewAxisHorizontal).
			BackgroundColor(ColorCardBody).
			Padding(Padding{}.All(L8)).
			Frame(Frame{Width: L480})
	})
}
