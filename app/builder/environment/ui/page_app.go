// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Ident: Custom-License

package uienv

import (
	"go.wdy.de/nago/app/builder/aam"
	"go.wdy.de/nago/app/builder/aam/nagogen"
	"go.wdy.de/nago/app/builder/environment"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/dataview"
)

func PageApp(wnd core.Window, uc environment.UseCases, ucAam aam.UseCases, ucGen nagogen.UseCases) core.View {
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

	initalText := core.AutoState[string](wnd)
	createPresented := core.AutoState[bool](wnd)
	createFn := core.AutoState[func(string)](wnd)

	return ui.VStack(
		dialogInputText(wnd, initalText.Get(), createPresented, createFn),
		ui.H1(uApp.Name+" ("+rstring.ActionEdit.Get(wnd)+")"),

		ui.HStack(
			ui.PrimaryButton(func() {
				zip, err := ucGen.Download(wnd.Subject(), model)
				if err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}

				wnd.ExportFiles(core.ExportFilesOptions{
					Files: []core.File{zip},
				})
			}).Title("Download"),
		).FullWidth(),
		ui.HStack(
			ui.TextField("App-ID", string(model.ID)).FullWidth().Disabled(true),
			ui.SecondaryButton(func() {
				initalText.Set(string(model.ID))
				createPresented.Set(true)
				createFn.Set(func(s string) {
					putEvent(environment.AppIDUpdated{ID: core.ApplicationID(s)})
				})
			}).PreIcon(icons.Edit),
		).FullWidth().Alignment(ui.Bottom).Gap(ui.L8),

		ui.HStack(
			ui.TextField("Git-Repo-URL", string(model.GitRepoURL)).FullWidth().Disabled(true),
			ui.SecondaryButton(func() {
				initalText.Set(string(model.GitRepoURL))
				createPresented.Set(true)
				createFn.Set(func(s string) {
					putEvent(environment.GitRepoUpdated{URL: core.URI(s)})
				})
			}).PreIcon(icons.Edit),
		).FullWidth().Alignment(ui.Bottom).Gap(ui.L8),

		ui.H2(StrNamespaces.Get(wnd)),
		dataview.FromData(wnd, dataview.Data[*aam.Namespace, environment.Ident]{
			FindAll:  model.Namespaces.Identifiers(),
			FindByID: model.Namespaces.FindByID,
			Fields: []dataview.Field[*aam.Namespace]{
				{
					ID:   "name",
					Name: "Name",
					Map: func(obj *aam.Namespace) core.View {
						return ui.Text(string(obj.Name))
					},
				},
			},
		}).CreateAction(func() {
			initalText.Set("")
			createPresented.Set(true)
			createFn.Set(func(v string) {
				putEvent(environment.NamespaceCreated{Name: environment.Ident(v)})
			})
		}).SelectOptions(
			dataview.NewSelectOptionDelete(wnd, func(selected []environment.Ident) error {
				for _, id := range selected {
					putEvent(environment.NamespaceDeleted{
						Name: id,
					})
				}

				return nil
			}),
		).NextActionIndicator(true).
			Action(func(e *aam.Namespace) {
				wnd.Navigation().ForwardTo(PathNamespace, wnd.Values().Put("namespace", string(e.Name)))
			}),
	).FullWidth().Alignment(ui.Leading).Gap(ui.L16)
}

func dialogInputText(wnd core.Window, initialText string, presented *core.State[bool], fn *core.State[func(string)]) core.View {
	if !presented.Get() {
		return nil
	}

	text := core.DerivedState[string](presented, "input").Init(func() string {
		return initialText
	})

	return alert.Dialog(
		rstring.ActionNew.Get(wnd),
		ui.TextField("", text.Get()).InputValue(text),
		presented,
		alert.Closeable(),
		alert.Large(),
		alert.Cancel(nil),
		alert.Create(func() (close bool) {
			fn.Get()(text.Get())
			return true
		}),
	)
}
