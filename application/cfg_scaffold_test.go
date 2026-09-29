// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application_test

import (
	"fmt"
	"slices"
	"testing"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
)

// TestScaffoldSubmenuVisibility ensures that the visibility rules of a sub menu entry are its own and not the
// ones of its parent entry.
func TestScaffoldSubmenuVisibility(t *testing.T) {
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.scaffoldtest")
		cfg.SetDecorator(cfg.NewScaffold().
			SubmenuEntry(func(menu *application.SubMenuBuilder) {
				menu.Title("parent")
				menu.MenuEntry().Title("everyone").Forward("page").Public()
				menu.MenuEntry().Title("private").Forward("page").Private()
				menu.MenuEntry().Title("permitted").Forward("page").OneOf(user.PermDelete)
				menu.MenuEntry().Title("public only").Forward("page").PublicOnly()
			}).
			Decorator())

		cfg.RootView("page", cfg.DecorateRootView(func(wnd core.Window) core.View {
			return ui.Text("page")
		}))
	})

	submenu := func(w *nagotest.Window) []string {
		t.Helper()
		scaffold := w.Find(nagotest.Type[*proto.Scaffold]()).Node().Component.(*proto.Scaffold)
		for _, entry := range scaffold.Menu {
			if entry.Title == "parent" {
				var titles []string
				for _, sub := range entry.Menu {
					titles = append(titles, string(sub.Title))
				}
				return titles
			}
		}

		t.Fatal("parent entry not found")
		return nil
	}

	anon := submenu(app.Open(t, nil, "page"))
	if !slices.Equal(anon, []string{"everyone", "public only"}) {
		t.Fatalf("unexpected anonymous sub menu %v", anon)
	}

	su := submenu(app.Open(t, user.SU(), "page"))
	if !slices.Equal(su, []string{"everyone", "private", "permitted"}) {
		t.Fatalf("unexpected authenticated sub menu %v", su)
	}
}

// TestScaffoldMenuBarVisible ensures that pages know already while being built, whether the scaffold shows the
// navigation bar or the burger menu, exactly at the breakpoint the frontend uses.
func TestScaffoldMenuBarVisible(t *testing.T) {
	configure := func(breakpoint int) func(cfg *application.Configurator) {
		return func(cfg *application.Configurator) {
			cfg.SetApplicationID("de.worldiety.scaffoldtest")
			scaffold := cfg.NewScaffold()
			if breakpoint > 0 {
				scaffold = scaffold.Breakpoint(breakpoint)
			}
			cfg.SetDecorator(scaffold.Decorator())

			cfg.RootView("page", cfg.DecorateRootView(func(wnd core.Window) core.View {
				return ui.Text(fmt.Sprintf("bar: %v class: %d", ui.ScaffoldMenuBarVisible(wnd), wnd.Info().SizeClass.Ordinal()))
			}))
		}
	}

	tests := []struct {
		breakpoint int
		width      int
		want       string
	}{
		{width: 767, want: "bar: false class: 1"},
		{width: 768, want: "bar: true class: 2"},
		{width: 1023, want: "bar: true class: 2"},
		{breakpoint: 1024, width: 1023, want: "bar: false class: 2"},
		{breakpoint: 1024, width: 1024, want: "bar: true class: 3"},
	}

	apps := map[int]*nagotest.App{
		0:    nagotest.New(t, configure(0)),
		1024: nagotest.New(t, configure(1024)),
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d/%d", tt.breakpoint, tt.width), func(t *testing.T) {
			w := apps[tt.breakpoint].Open(t, nil, "page", nagotest.Size(tt.width, 800))
			w.Find(nagotest.Text(tt.want))

			// the frontend switches at the breakpoint sent with the scaffold, zero meaning its default
			scaffold := w.Find(nagotest.Type[*proto.Scaffold]()).Node().Component.(*proto.Scaffold)
			if int(scaffold.Breakpoint) != tt.breakpoint {
				t.Fatalf("expected breakpoint %d, got %d", tt.breakpoint, scaffold.Breakpoint)
			}
		})
	}
}
