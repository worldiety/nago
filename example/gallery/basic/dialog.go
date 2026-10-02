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
	"go.wdy.de/nago/presentation/ui/alert"
)

func init() {
	app.Register("dialog", func(wnd core.Window) core.View {
		presented := core.AutoState[bool](wnd)

		return VStack(
			alert.Dialog(
				"Delete project",
				Text("Do you really want to delete the project? This cannot be undone."),
				presented,
				alert.Closeable(),
				alert.Cancel(nil),
				alert.Delete(func() {
					// delete the project
				}),
			),
			PrimaryButton(func() { presented.Set(true) }).Title("Delete project"),
		)
	})
}
