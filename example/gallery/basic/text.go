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
	app.Register("text", func(wnd core.Window) core.View {
		return VStack(
			Text("Hello Nago").Font(Title),
			Text("A plain paragraph of text."),
			Text("Colored and underlined").Color(SE0).Underline(true),
			Link(wnd, "A link to nago.dev", "https://www.nago.dev", "_blank"),
			MailTo(wnd, "Write us a mail", "info@example.com"),
		).Alignment(Leading).Gap(L8)
	})
}
