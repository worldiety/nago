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
	"go.wdy.de/nago/presentation/ui/breadcrumb"
)

func init() {
	app.Register("breadcrumbs", func(wnd core.Window) core.View {
		return breadcrumb.Breadcrumbs().
			ClampLeading().
			Item("Home", func() { wnd.Navigation().ForwardTo("/", nil) }).
			Item("Projects", func() {}).
			Item("Nago", func() {})
	})
}
