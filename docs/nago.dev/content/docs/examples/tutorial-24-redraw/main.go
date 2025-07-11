// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	_ "embed"
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"sync/atomic"
	"time"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		var redrawCounter atomic.Uint64
		cfg.RootView(".", func(wnd core.Window) core.View {
			// this shows how to redraw views with a fixed rate, without using states to signal view invalidation
			redrawCounter.Add(1)

			return ui.RedrawAtFixedRate(wnd, time.Second, ui.Text(fmt.Sprintf("redraw: %v", redrawCounter.Load())))
		})
	}).Run()
}
