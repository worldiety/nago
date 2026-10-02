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
	app.Register("rich-text", func(wnd core.Window) core.View {
		return RichText(`<h2>Release notes</h2>
<p>This release brings <b>faster rendering</b> and <i>many</i> small fixes.</p>
<ul>
  <li>New slider component</li>
  <li>Improved dialogs</li>
</ul>`).Frame(Frame{Width: L480})
	})
}
