// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uienv

import (
	"fmt"
	"os"

	"go.wdy.de/nago/app/builder/aam"
	"go.wdy.de/nago/app/builder/environment"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/dataview"
)

func PageNamespace(wnd core.Window, uc environment.UseCases, ucAam aam.UseCases) core.View {
	uApp, errV := loadApp(wnd, uc)
	if errV != nil {
		return errV
	}

	env := environment.ID(wnd.Values()["env"])

	putEvent := func(evt environment.Event) {
		if err := uc.PutEvent(wnd.Subject(), env, uApp.ID, evt); err != nil {
			alert.ShowBannerError(wnd, err)
		}
	}

	model, err := ucAam.Create(wnd.Subject(), env, uApp.ID)
	if err != nil {
		return alert.BannerError(err)
	}
	ns, ok := model.Namespaces.Get(environment.Ident(wnd.Values()["namespace"]))
	if !ok {
		return alert.BannerError(fmt.Errorf("namespace not found: %s: %w", wnd.Values()["namespace"], os.ErrNotExist))
	}

	initalText := core.AutoState[string](wnd)
	createPresented := core.AutoState[bool](wnd)
	createFn := core.AutoState[func(string)](wnd)

	return ui.VStack(
		dialogInputText(wnd, initalText.Get(), createPresented, createFn),
		ui.H1(uApp.Name+" ("+rstring.ActionEdit.Get(wnd)+")"),
		ui.HStack(),
		ui.H2("Structs"),
		dataview.FromData(wnd, dataview.Data[*aam.Struct, environment.Ident]{
			FindAll:  ns.Structs.Identifiers(),
			FindByID: ns.Structs.FindByID,
			Fields: []dataview.Field[*aam.Struct]{
				{
					ID:   "name",
					Name: "Name",
					Map: func(obj *aam.Struct) core.View {
						return ui.Text(string(obj.Name))
					},
				},
			},
		}).CreateAction(func() {
			initalText.Set("")
			createPresented.Set(true)
			createFn.Set(func(v string) {
				putEvent(environment.StructCreated{
					Namespace: ns.Name,
					Name:      environment.Ident(v),
				})
			})
		}).SelectOptions(
			dataview.NewSelectOptionDelete(wnd, func(selected []environment.Ident) error {
				for _, id := range selected {
					putEvent(environment.TypeDeleted{
						Namespace: ns.Name,
						Name:      id,
					})
				}

				return nil
			}),
		).NextActionIndicator(true),
	).FullWidth().Alignment(ui.Leading).Gap(ui.L16)
}
