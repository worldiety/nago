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
	"go.wdy.de/nago/presentation/ui/tags"
)

func init() {
	app.Register("colored-text-pill", func(wnd core.Window) core.View {
		return VStack(
			HStack(
				tags.ColoredTextPill(SG0, "Done"),
				tags.ColoredTextPill(SW0, "In review"),
				tags.ColoredTextPill(SE0, "Rejected"),
			).Gap(L8),
			HStack(
				tags.StatusBadge(SG0, "Online"),
				tags.StatusBadge(SE0, "Failed"),
				tags.StatusBadge("", "Unknown"),
			).Gap(L8),
		).Alignment(Leading).Gap(L16)
	})
}
