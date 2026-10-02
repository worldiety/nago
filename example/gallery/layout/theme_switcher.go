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
	app.Register("theme-switcher", func(wnd core.Window) core.View {
		return ThemeSwitcher(SecondaryButton(nil).PreIcon(icons.Swatch).Title("Theme"))
	})
}
