// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/crud"
	"go.wdy.de/nago/web/vuejs"
)

type Invoice struct {
	ID string
	A  int
	B  int
}

func (i Invoice) Identity() string {
	return i.ID
}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {
			bnd := crud.NewBinding[Invoice](wnd)
			bnd.Add(
				crud.Int(crud.IntOptions{Label: "A"}, crud.Ptr(func(model *Invoice) *int {
					return &model.A
				})),
				crud.Int(crud.IntOptions{Label: "B"}, crud.Ptr(func(model *Invoice) *int {
					return &model.B
				})),
				crud.Text(crud.TextOptions{Label: "A+B"}, crud.Ptr(func(model *Invoice) *string {
					tmp := fmt.Sprintf("%.2f", float64(model.A+model.B))
					return &tmp
				})),
			)

			state := core.AutoState[Invoice](wnd)
			return ui.VStack(
				crud.Form(bnd, state)).
				Frame(ui.Frame{}.MatchScreen())
		})
	}).Run()
}
