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
	app.Register("three-column", func(wnd core.Window) core.View {
		return navsplitview.ThreeColumn(navsplitview.NavLinks{
			"folders": VStack(
				navsplitview.ListItem(navsplitview.KindContent, "work", Text("Work")).
					DeleteTarget(navsplitview.KindDetail),
				navsplitview.ListItem(navsplitview.KindContent, "private", Text("Private")).
					DeleteTarget(navsplitview.KindDetail),
			).FullWidth(),
			"none": Text("Nothing selected"),
			"work": VStack(
				navsplitview.ListItem(navsplitview.KindDetail, "report", Text("Report.pdf")),
			).FullWidth(),
			"private": VStack(
				navsplitview.ListItem(navsplitview.KindDetail, "photo", Text("Photo.jpg")),
			).FullWidth(),
			"report": Text("Quarterly report"),
			"photo":  Text("Holiday photo"),
		}).Default("folders", "work", "none").
			WidthSidebar(L160).
			WidthContent(L200).
			BackgroundColorSidebar(ColorCardBody).
			Frame(Frame{Width: L880, Height: L160})
	})
}
