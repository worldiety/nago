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
	"go.wdy.de/nago/presentation/ui/avatar"
	"go.wdy.de/nago/presentation/ui/list"
)

func init() {
	app.Register("list", func(wnd core.Window) core.View {
		person := func(name, role string) list.TEntry {
			return list.Entry().
				Leading(avatar.Text(name)).
				Headline(name).
				SupportingText(role).
				Trailing(ImageIcon(icons.ChevronRight)).
				Action(func() {
					// navigate to the details of the person
				})
		}

		return list.List(
			person("Ada Lovelace", "Engineering"),
			person("Grace Hopper", "Operations"),
			person("Alan Turing", "Research"),
		).
			Caption(Text("Team members")).
			Footer(Text("3 entries")).
			Frame(Frame{Width: L400})
	})
}
