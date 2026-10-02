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
	"go.wdy.de/nago/presentation/ui/navsplitview"
)

func init() {
	app.Register("two-column", func(wnd core.Window) core.View {
		return navsplitview.TwoColumn(navsplitview.NavLinks{
			"list": VStack(
				navsplitview.ListItem(navsplitview.KindDetail, "inbox", Text("Inbox")),
				navsplitview.ListItem(navsplitview.KindDetail, "sent", Text("Sent")),
			).FullWidth(),
			"none":  Text("Nothing selected"),
			"inbox": Text("3 new messages"),
			"sent":  Text("No messages sent yet"),
		}).Default("list", "none").
			WidthContent(L200).
			BackgroundColorContent(ColorCardBody).
			Frame(Frame{Width: L560, Height: L160})
	})
}
