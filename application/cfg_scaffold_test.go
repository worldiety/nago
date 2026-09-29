// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application_test

import (
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
