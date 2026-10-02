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
)

func init() {
	app.Register("menu", func(wnd core.Window) core.View {
		item := func(icon core.SVG, title string) TMenuItem {
			return MenuItem(func() {
				// perform the action
			}, HStack(ImageIcon(icon), Text(title)).Gap(L8))
		}

		return VStack(
			Menu(
				SecondaryButton(nil).Title("Actions").PostIcon(icons.ChevronDown),
				MenuGroup(
					item(icons.PencilSquare, "Edit"),
					item(icons.DocumentDuplicate, "Duplicate"),
				),
				MenuGroup(
					item(icons.Trash, "Delete"),
				),
			),
		).Alignment(TopLeading).Frame(Frame{Height: L256, Width: L320})
	})
}
