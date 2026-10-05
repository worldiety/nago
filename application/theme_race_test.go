// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/theme"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// The theme may be updated by any goroutine, also while the application is built and while windows render with
// its colors. Run with -race.
func TestThemeUpdatesWhileBuildingAndRendering(t *testing.T) {
	var updates sync.WaitGroup
	var stop atomic.Bool
	var last atomic.Value
	var themes application.ThemeManagement

	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.themerace")
		themes = std.Must(cfg.ThemeManagement())
		cfg.RootView(".", func(wnd core.Window) core.View {
			return ui.Text(string(core.Colors[ui.Colors](wnd).M0))
		})

		// a goroutine of the application, like a scheduler, updates the theme while the application is built
		updates.Add(1)
		go func() {
			defer updates.Done()
			for n := 0; !stop.Load() && n < 2000; n++ {
				base := theme.DefaultBaseColors
				base.Main = ui.Color(fmt.Sprintf("#%02x8c30", n%256))
				colors := theme.Colors{Dark: theme.DarkMode(base), Light: theme.LightMode(base)}
				if err := themes.UseCases.UpdateColors(user.SU(), colors); err != nil {
					t.Error(err)
					return
				}
				last.Store(colors)
			}
		}()
	})

	for range 20 {
		app.Open(t, nil, ".")
	}

	stop.Store(true)
	updates.Wait()

	// no update got lost: the bus delivers asynchronously, but the application ends up with the last stored colors
	want := last.Load().(theme.Colors)
	deadline := time.Now().Add(5 * time.Second)
	for {
		w := app.Open(t, nil, ".")
		if w.FindAll(nagotest.Text(string(want.Dark.M0))).Len()+w.FindAll(nagotest.Text(string(want.Light.M0))).Len() > 0 {
			return
		}

		if time.Now().After(deadline) {
			t.Fatalf("the application does not show the last colors %s or %s", want.Dark.M0, want.Light.M0)
		}

		time.Sleep(10 * time.Millisecond)
	}
}
