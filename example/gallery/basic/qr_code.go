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
	app.Register("qr-code", func(wnd core.Window) core.View {
		return QrCode("https://www.nago.dev").
			AccessibilityLabel("QR code linking to nago.dev").
			Frame(Frame{}.Size(L200, L200))
	})
}
