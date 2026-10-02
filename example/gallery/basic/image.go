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
	icons "go.wdy.de/nago/presentation/icons/hero/solid"
	. "go.wdy.de/nago/presentation/ui"
)

const logo = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 120 80">
<rect width="120" height="80" rx="12" fill="#1b8c30"/>
<circle cx="40" cy="40" r="20" fill="#ffffff"/>
<rect x="68" y="22" width="32" height="36" rx="6" fill="#ffffff"/>
</svg>`

func init() {
	app.Register("image", func(wnd core.Window) core.View {
		return HStack(
			Image().
				Embed([]byte(logo)).
				AccessibilityLabel("A green logo").
				Frame(Frame{}.Size(L160, L120)),
			ImageIcon(icons.Heart),
			ImageIcon(icons.Star).FillColor(SW0),
			ImageIcon(icons.Bell).Frame(Frame{}.Size(L48, L48)),
		).Gap(L24)
	})
}
