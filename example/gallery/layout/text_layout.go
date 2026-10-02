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
	app.Register("text-layout", func(wnd core.Window) core.View {
		return TextLayout(
			Text("TextLayout flows "),
			Text("bold").Font(Font{Weight: HeadlineAndTitleFontWeight}),
			Text(", "),
			Text("monospaced").Font(Monospace),
			Text(" and "),
			Text("colored").Color("#FA2C7F"),
			Text(" text inline, like a paragraph, and wraps it at the end of the line."),
		).Frame(Frame{Width: L320})
	})
}
