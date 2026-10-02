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
	"go.wdy.de/nago/presentation/ui/markdown"
)

func init() {
	app.Register("markdown", func(wnd core.Window) core.View {
		return VStack(
			markdown.RichText("Nago renders **bold** and *italic* text:\n\n- lists\n- [links](https://www.nago.dev)"),
			markdown.Render(markdown.Options{Window: wnd}, []byte("## Release notes\n\nRead the [documentation](https://www.nago.dev).")),
		).Alignment(Leading).Gap(L32).Frame(Frame{Width: L480})
	})
}
