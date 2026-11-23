// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uitheme_test

import (
	"strings"
	"testing"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/theme"
	cfgtheme "go.wdy.de/nago/application/theme/cfg"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
)

var (
	handmadeLightBase = theme.BaseColors{Main: "#033303", Interactive: "#033303", Accent: "#033303"}
	handmadeDarkBase  = theme.BaseColors{Main: "#033303", Interactive: "#f7a823", Accent: "#f7a823"}
)

// handmade are colors which an application proposes after its design manual: some of them are not calculated from
// the base colors, so the page keeps them as adjustments.
func handmade() theme.Colors {
	c := theme.Colors{Light: theme.LightMode(handmadeLightBase), Dark: theme.DarkMode(handmadeDarkBase)}
	c.Light.M1, c.Light.Error = "#f1f0eb", "#bf2d28"
	c.Dark.M1, c.Dark.Error = "#0a1f0a", "#e5484d"
	return c
}

// colorField selects the hex value of the color field with the given label, which opens its picker on a click.
func colorField(label string) nagotest.Matcher {
	return nagotest.Where("color field "+label, func(n nagotest.Node) bool {
		hex, ok := n.Component.(*proto.TextView)
		if !ok || !strings.HasPrefix(string(hex.Value), "#") || len(n.Parents) < 2 {
			return false
		}

		outer, ok := n.Parents[len(n.Parents)-2].(*proto.Stack)
		if !ok || len(outer.Children) == 0 {
			return false
		}

		title, ok := outer.Children[0].(*proto.TextView)
		return ok && string(title.Value) == label
	})
}

// pickMain picks the given main color in the dialog of the field and applies it.
func pickMain(w *nagotest.Window, color ui.Color) {
	w.Click(w.Find(colorField("Hauptfarbe")))
	w.Type(w.Find(nagotest.Type[*proto.ColorPicker]()), string(color))
	w.Click(w.Find(nagotest.Text("Übernehmen")))
}

// save clicks the save button of the page, which is the last one, as the preview shows a button of the same name.
func save(t *testing.T, w *nagotest.Window) {
	t.Helper()
	all := w.FindAll(nagotest.Text("Speichern"))
	if all.Len() == 0 {
		t.Fatal("no save button")
	}

	w.Click(all.At(all.Len() - 1))
}

// In the simple mode, a changed base color calculates all colors of its mode again, because the user only sees the
// base colors. In the advanced mode, the adjusted colors stay, because the user made them on purpose. The other
// mode stays as it is.
func TestBaseColorsInSimpleAndAdvancedMode(t *testing.T) {
	const picked = ui.Color("#1d4ed8")
	for _, row := range []struct {
		name     string
		advanced bool
		dark     bool
		act      func(w *nagotest.Window)
		base     theme.BaseColors // expected base colors of the edited mode
	}{
		{
			name: "simple mode, main color of the light mode",
			act:  func(w *nagotest.Window) { pickMain(w, picked) },
			base: theme.BaseColors{Main: picked, Interactive: handmadeLightBase.Interactive, Accent: handmadeLightBase.Accent},
		},
		{
			name:     "advanced mode, main color of the light mode",
			advanced: true,
			act:      func(w *nagotest.Window) { pickMain(w, picked) },
			base:     theme.BaseColors{Main: picked, Interactive: handmadeLightBase.Interactive, Accent: handmadeLightBase.Accent},
		},
		{
			name: "simple mode, base colors of the light mode copied into the dark mode",
			dark: true,
			act:  func(w *nagotest.Window) { w.Click(w.Find(nagotest.Text("Basisfarben vom Light-Mode übernehmen"))) },
			base: handmadeLightBase,
		},
		{
			name:     "advanced mode, base colors of the light mode copied into the dark mode",
			advanced: true,
			dark:     true,
			act:      func(w *nagotest.Window) { w.Click(w.Find(nagotest.Text("Basisfarben vom Light-Mode übernehmen"))) },
			base:     handmadeLightBase,
		},
	} {
		t.Run(row.name, func(t *testing.T) {
			var themes application.ThemeManagement
			app := nagotest.New(t, func(cfg *application.Configurator) {
				cfg.SetApplicationID("de.worldiety.theme.colors")
				cfg.SetThemeDefaults(handmade())
				themes = std.Must(cfg.ThemeManagement())
				std.Must(cfgtheme.Enable(cfg))
			})

			proposed := handmade()
			before, err := themes.UseCases.ReadCustomization(user.SU())
			if err != nil {
				t.Fatal(err)
			}

			// premise: the page keeps the colors of the design manual as adjustments in both modes
			for _, sc := range []theme.SchemeCustomization{before.Light, before.Dark} {
				if sc.Overrides["M1"] == "" || sc.Overrides["SE0"] == "" {
					t.Fatalf("premise: the proposal is not kept as adjustments: %v", sc.Overrides)
				}
			}

			w := app.Open(t, user.SU(), "admin/theme/colors")
			if row.advanced {
				w.Type(w.Find(nagotest.Type[*proto.Toggle]()), "true")
			}
			if row.dark {
				w.Click(w.Find(nagotest.Text("Dark-Mode")))
			}

			row.act(w)
			save(t, w)

			after, err := themes.UseCases.ReadCustomization(user.SU())
			if err != nil {
				t.Fatal(err)
			}

			colors, err := themes.UseCases.ReadColors(user.SU())
			if err != nil {
				t.Fatal(err)
			}

			edited, other := after.Light, after.Dark
			editedColors, editedProposal, derive := colors.Light, proposed.Light, theme.LightMode
			otherBefore := before.Dark
			if row.dark {
				edited, other = after.Dark, after.Light
				editedColors, editedProposal, derive = colors.Dark, proposed.Dark, theme.DarkMode
				otherBefore = before.Light
			}

			// in both modes, the action reaches the stored base colors
			if edited.Base != row.base {
				t.Fatalf("stored base colors %v, want %v", edited.Base, row.base)
			}

			calculated := derive(row.base)
			if row.advanced {
				if edited.Overrides["M1"] != editedProposal.M1 || edited.Overrides["SE0"] != editedProposal.Error {
					t.Errorf("advanced mode lost the adjustments: %v", edited.Overrides)
				}
				if editedColors.M1 != editedProposal.M1 || editedColors.Error != editedProposal.Error {
					t.Errorf("advanced mode: M1 %s and Error %s, want the adjusted %s and %s", editedColors.M1, editedColors.Error, editedProposal.M1, editedProposal.Error)
				}
			} else {
				if len(edited.Overrides) != 0 {
					t.Errorf("simple mode kept %d adjustments: %v", len(edited.Overrides), edited.Overrides)
				}
				if editedColors != calculated {
					t.Errorf("simple mode: M1 %s and Error %s, want the calculated %s and %s", editedColors.M1, editedColors.Error, calculated.M1, calculated.Error)
				}
			}

			// the other mode stays as it is
			if other.Base != otherBefore.Base || len(other.Overrides) != len(otherBefore.Overrides) || other.Overrides["M1"] != otherBefore.Overrides["M1"] {
				t.Errorf("the other mode changed: %v, before %v", other, otherBefore)
			}
		})
	}
}
