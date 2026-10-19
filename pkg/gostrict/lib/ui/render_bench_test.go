// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui_test

import (
	"context"
	"testing"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/gostrict/engine/vm"
	"go.wdy.de/nago/pkg/gostrict/lib/ui"
	"go.wdy.de/nago/pkg/gostrict/lib/ui/internal/form"
	"go.wdy.de/nago/presentation/core"
)

// BenchmarkRender measures a render of the form, which a change of a field triggers, with many views and thus many
// calls of the functions of lib/ui.
func BenchmarkRender(b *testing.B) {
	in := loadDir(b, "internal/form")
	inVM := loadWith(b, "internal/form", new(vm.Engine))
	for name, render := range map[string]func(wnd core.Window) core.View{
		"native": func(wnd core.Window) core.View { return ui.Core(form.Render(ui.WindowOf(wnd))) },
		"goja":   func(wnd core.Window) core.View { return ui.Render(context.Background(), in, "Render", wnd) },
		"vm":     func(wnd core.Window) core.View { return ui.Render(context.Background(), inVM, "Render", wnd) },
	} {
		b.Run(name, func(b *testing.B) {
			app := nagotest.New(b, func(cfg *application.Configurator) {
				cfg.SetApplicationID(core.ApplicationID("de.worldiety.gostrict.render." + name))
				cfg.RootView("form", render)
			})
			w := app.Open(b, nil, "form")
			field := w.Find(nagotest.Label("Name"))
			i := 0
			for b.Loop() {
				i++
				if i%2 == 0 {
					w.Type(field, "Anna")
				} else {
					w.Type(field, "Bo")
				}
			}
		})
	}
}
