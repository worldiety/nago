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
	icons "go.wdy.de/nago/presentation/icons/hero/outline"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/editor"
	"go.wdy.de/nago/presentation/ui/list"
)

func init() {
	app.Register("screen", func(wnd core.Window) core.View {
		pagesVisible := core.AutoState[bool](wnd).Init(func() bool { return true })

		pages := editor.ToolWindow(icons.DocumentText, "Pages").
			Content(VStack(
				list.List(
					list.Entry().Headline("Home"),
					list.Entry().Headline("About us"),
					list.Entry().Headline("Contact"),
				).FullWidth(),
			).FullWidth().Padding(Padding{}.All(L8))).
			Visible(pagesVisible.Get())

		return editor.Screen("Pages").
			Header(editor.Header(wnd).Center(Text("Edit page"))).
			Navbar(editor.Navbar().
				Top(TertiaryButton(func() {
					pagesVisible.Set(!pagesVisible.Get())
				}).PreIcon(icons.DocumentText)).
				Bottom(TertiaryButton(nil).PreIcon(icons.Cog6Tooth)),
			).
			TrailingToolWindow(pages).
			Content(editor.Content(
				VStack(
					Text("About us").Font(Title),
					Text("We build business software in Go."),
				).Alignment(Leading).Gap(L16),
			).Style(editor.ContentPage))
	})
}
