// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Ident: Custom-License

package uienv

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/app/builder/app"
	"go.wdy.de/nago/app/builder/environment"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/dataview"
	"go.wdy.de/nago/presentation/ui/form"
)

func PageEnvironments(wnd core.Window, uc environment.UseCases) core.View {
	envs, err := xslices.Collect2(uc.FindAll(wnd.Subject()))
	if err != nil {
		return alert.BannerError(err)
	}

	createEnvPresented := core.AutoState[bool](wnd)
	createAppPresented := core.AutoState[bool](wnd)
	selectedEnv := core.AutoState[environment.ID](wnd)

	return ui.VStack(
		dialogNewEnvironment(wnd, uc, createEnvPresented),
		dialogNewApp(wnd, uc, selectedEnv, createAppPresented),
		ui.H1(StrAppsAndEnvironments.Get(wnd)),
		ui.HStack(
			ui.SecondaryButton(func() {
				createEnvPresented.Set(true)
			}).Title(StrCreateEnvironment.Get(wnd)),
		).FullWidth().Alignment(ui.Trailing),
	).Append(
		ui.ForEach(envs, func(env environment.Environment) core.View {
			return ui.VStack(
				ui.H2(env.Name),
				ui.H2(env.Description),
				dataview.FromData(wnd, dataview.Data[app.App, app.ID]{
					FindAll: xslices.ValuesWithError(env.Apps, nil),
					FindByID: func(id app.ID) (option.Opt[app.App], error) {
						return uc.FindAppByID(wnd.Subject(), env.ID, id)
					},
					Fields: []dataview.Field[app.App]{
						{
							ID:   "name",
							Name: rstring.LabelName.Get(wnd),
							Map: func(obj app.App) core.View {
								return ui.Text(obj.Name)
							},
						},
					},
				}).
					Action(func(e app.App) {
						wnd.Navigation().ForwardTo(PathApp, wnd.Values().Put("env", string(env.ID)).Put("app", string(e.ID)))
					}).
					NextActionIndicator(true).
					CreateAction(func() {
						selectedEnv.Set(env.ID)
						createAppPresented.Set(true)
					}),
			).FullWidth().Alignment(ui.Leading)
		})...,
	).FullWidth().Alignment(ui.Leading)
}

func dialogNewEnvironment(wnd core.Window, uc environment.UseCases, presented *core.State[bool]) core.View {
	if !presented.Get() {
		return nil
	}

	formState := core.AutoState[environment.CreateOptions](wnd)

	return alert.Dialog(StrCreateEnvironment.Get(wnd), form.Auto(form.AutoOptions{}, formState), presented, alert.Closeable(), alert.Cancel(nil), alert.Create(func() (close bool) {
		if _, err := uc.Create(wnd.Subject(), formState.Get()); err != nil {
			alert.ShowBannerError(wnd, err)
			return
		}

		return true
	}))
}

func dialogNewApp(wnd core.Window, uc environment.UseCases, env *core.State[environment.ID], presented *core.State[bool]) core.View {
	if !presented.Get() {
		return nil
	}

	formState := core.AutoState[environment.CreateAppOptions](wnd)

	return alert.Dialog(StrCreateApp.Get(wnd), form.Auto(form.AutoOptions{}, formState), presented, alert.Closeable(), alert.Cancel(nil), alert.Create(func() (close bool) {
		if _, err := uc.CreateApp(wnd.Subject(), env.Get(), formState.Get()); err != nil {
			alert.ShowBannerError(wnd, err)
			return
		}

		return true
	}))
}
