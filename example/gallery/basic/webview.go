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
	"go.wdy.de/nago/presentation/ui/webview"
)

func init() {
	app.Register("webview", func(wnd core.Window) core.View {
		return webview.WebView().
			Title("Embedded page").
			Raw(`<html><body style="font-family: sans-serif; background: #e8f0fe; padding: 16px">
<h3>Embedded HTML</h3>
<p>This content is isolated in an iframe.</p>
</body></html>`).
			Frame(Frame{Width: L400, Height: L200})
	})
}
