// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"fmt"

	"go.wdy.de/nago/example/gallery/app"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/alert"
)

func init() {
	app.Register("banner-error", func(wnd core.Window) core.View {
		err := fmt.Errorf("cannot connect to database: connection refused")

		// the user only sees a generic message, the details go into the log
		return alert.BannerError(err)
	})
}
