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
	"go.wdy.de/nago/presentation/ui/progress"
)

func init() {
	app.Register("progress", func(wnd core.Window) core.View {
		return VStack(
			Text("Uploading 3 of 4 files"),
			progress.LinearProgress().Progress(0.75),
			Text("Storage used"),
			progress.LinearProgress().Progress(0.3).Color(SE0),
		).Alignment(Leading).Gap(L8).Frame(Frame{Width: L400})
	})
}
