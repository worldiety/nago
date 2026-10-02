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
	app.Register("banner", func(wnd core.Window) core.View {
		presented := core.AutoState[bool](wnd).Init(func() bool { return true })

		return VStack(
			alert.Banner("Saved", "Your changes have been saved.").
				Intent(alert.IntentSuccess).
				Closeable(presented),
			alert.Banner("Note", "The system will be updated tonight.").
				Intent(alert.IntentOk),
			alert.Banner("Quota", "You use 90% of your storage.").
				Intent(alert.IntentWarning),
			alert.Banner("Failed", "The file could not be uploaded.").
				Intent(alert.IntentError),
		).Gap(L16)
	})
}
