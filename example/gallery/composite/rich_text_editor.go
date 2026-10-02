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
	app.Register("rich-text-editor", func(wnd core.Window) core.View {
		html := core.AutoState[string](wnd).Init(func() string {
			return "<h2>Release notes</h2><p>This release brings <b>faster</b> start-up and a new <i>dark mode</i>.</p><ul><li>Improved search</li><li>New export</li></ul>"
		})

		return RichTextEditor(html.Get()).
			InputValue(html).
			Frame(Frame{Width: L560})
	})
}
