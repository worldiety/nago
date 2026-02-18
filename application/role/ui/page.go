// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uirole

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/application/rcrud"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/dataview"
)

type Pages struct {
	Roles core.NavigationPath
}

func GroupPage(wnd core.Window, useCases rcrud.UseCases[role.Role, role.ID]) core.View {
	createPresented := core.AutoState[bool](wnd)
	createWithCustomID := core.AutoState[bool](wnd)

	return ui.VStack(
		ui.H1(rstring.LabelRoles.Get(wnd)),
		dataview.FromData(wnd, dataview.Data[role.Role, role.ID]{
			FindAll: func(yield func(role.ID, error) bool) {
				for r, err := range useCases.FindAll(wnd.Subject()) {
					if !yield(r.ID, err) {
						return
					}
				}
			},
			FindByID: func(id role.ID) (option.Opt[role.Role], error) {
				return useCases.FindByID(wnd.Subject(), id)
			},
			Fields: []dataview.Field[role.Role]{
				{
					ID:   "name",
					Name: rstring.LabelName.Get(wnd),
					Map: func(obj role.Role) core.View {
						return ui.Text(obj.Name)
					},
				},
				{
					ID:   "description",
					Name: rstring.LabelDescription.Get(wnd),
					Map: func(obj role.Role) core.View {
						return ui.Text(obj.Description)
					},
				},
			},
		}).Search(true).
			NextActionIndicator(true).
			CreateOptions(
				dataview.CreateOption{
					Name: "Role with predefined ID",
					Action: func() error {
						createWithCustomID.Set(true)
						createPresented.Set(true)
						return nil
					},
				},
				dataview.CreateOption{
					Name: "Create Role",
					Action: func() error {
						createWithCustomID.Set(false)
						createPresented.Set(true)
						return nil
					},
				},
			).SelectOptions(
			dataview.NewSelectOptionDelete[role.ID](wnd, func(selected []role.ID) error {
				for _, id := range selected {
					if err := useCases.DeleteByID(wnd.Subject(), id); err != nil {
						return err
					}
				}

				return nil
			}),
		),
	).FullWidth().
		Alignment(ui.Leading)
}
