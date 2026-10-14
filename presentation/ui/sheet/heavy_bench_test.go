// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package sheet_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/sheet"
)

// heavyPage puts the sheet beside a lot of other views, like a form with many fields, so that a render of the
// whole page costs much more than the window of the sheet.
func heavyPage(others int) func(wnd core.Window) core.View {
	return func(wnd core.Window) core.View {
		var views []core.View
		for i := range others {
			state := core.StateOf[string](wnd, "field"+strconv.Itoa(i))
			views = append(views, ui.HStack(
				ui.Text(fmt.Sprintf("Feld %d", i)),
				ui.TextField("Wert", state.Get()).InputValue(state),
				ui.SecondaryButton(func() {}).Title("Aktion"),
			).Gap(ui.L8))
		}

		return ui.VStack(
			ui.VStack(views...),
			sheet.Sheet("big", 50_000, 30, func(row, col int) sheet.Cell {
				return sheet.Cell{Text: strconv.Itoa(row*30 + col), Type: sheet.ValueNumber}
			}).Frozen(1, 1),
		)
	}
}

// TestHeavyPageScroll measures a change of the window on pages of growing weight. It is skipped by default,
// run it with -run TestHeavyPageScroll -v -heavy.
func TestHeavyPageScroll(t *testing.T) {
	if !*heavy {
		t.Skip("run with -heavy")
	}

	for _, others := range []int{0, 100, 500, 2000} {
		app := nagotest.New(t, func(cfg *application.Configurator) {
			cfg.SetApplicationID("de.worldiety.sheetheavy")
			cfg.RootView("page", heavyPage(others))
		}, nagotest.WithRoundTrip())

		w := app.Open(t, nil, "page")
		s := w.Find(nagotest.Type[*proto.Sheet]())
		const n = 20
		start := time.Now()
		for i := range n {
			w.SheetScroll(s, sheet.Range{Row: 1 + i*500, Col: 1, Rows: 90, Cols: 29})
		}
		perScroll := time.Since(start) / n
		t.Logf("%5d other views: %v per window change", others, perScroll.Round(10*time.Microsecond))
	}
}
