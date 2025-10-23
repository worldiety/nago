// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiai

import (
	"fmt"
	"os"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/ai"
	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/dataview"
	"golang.org/x/text/language"
)

var (
	StrLibraries = i18n.MustString("nago.ai.admin.libraries", i18n.Values{language.English: "Libraries", language.German: "Bibliotheken"})
	StrAgents    = i18n.MustString("nago.ai.admin.agents", i18n.Values{language.English: "Agents", language.German: "Agenten"})
)

func PageProvider(wnd core.Window, uc ai.UseCases) core.View {
	optProv, err := uc.FindProviderByID(wnd.Subject(), provider.ID(wnd.Values()["provider"]))
	if err != nil {
		return alert.BannerError(err)
	}

	if optProv.IsNone() {
		return alert.BannerError(fmt.Errorf("provider not found: %s: %w", wnd.Values()["provider"], os.ErrNotExist))
	}

	prov := optProv.Unwrap()

	return ui.VStack(
		ui.H1(prov.Name()),
		ui.Text(prov.Description()),
		ui.Space(ui.L48),
		agentTable(wnd, prov),
		ui.Space(ui.L48),
		libTable(wnd, prov),
	).FullWidth().Alignment(ui.Leading)
}

func agentTable(wnd core.Window, prov provider.Provider) core.View {
	optAgents := prov.Agents()
	if optAgents.IsNone() {
		return nil
	}

	agents := optAgents.Unwrap()

	loadedAgents := core.AutoState[[]agent.Agent](wnd).AsyncInit(func() []agent.Agent {
		v, err := xslices.Collect2(agents.All(wnd.Subject()))
		if err != nil {
			alert.ShowBannerError(wnd, err)
			return nil
		}

		return v
	})

	return ui.VStack(
		ui.H2(StrAgents.Get(wnd)),
		ui.If(!loadedAgents.Valid(), ui.Text(rstring.LabelPleaseWait.Get(wnd))),

		ui.IfFunc(loadedAgents.Valid(), func() core.View {
			return dataview.FromSlice(wnd, loadedAgents.Get(), []dataview.Field[dataview.Element[agent.Agent]]{
				{
					Name: rstring.LabelName.Get(wnd),
					Map: func(obj dataview.Element[agent.Agent]) core.View {
						return ui.Text(obj.Value.Name)
					},
				},
			}).NextActionIndicator(true).
				ActionNew(func() {

				}).SelectOptions(
				dataview.NewSelectOptionDelete(wnd, func(selected []dataview.Idx) error {
					for _, i := range selected {
						if idx, ok := i.Int(); ok {
							if err := agents.Delete(wnd.Subject(), loadedAgents.Get()[idx].ID); err != nil {
								return err
							}
						}
					}

					loadedAgents.SetValid(false)

					return nil
				}),
			)
		}),
	).FullWidth().Alignment(ui.Leading)

}

func libTable(wnd core.Window, prov provider.Provider) core.View {
	optLibs := prov.Libraries()
	if optLibs.IsNone() {
		return nil
	}

	libs := optLibs.Unwrap()

	loadedLibs := core.AutoState[[]library.Library](wnd).AsyncInit(func() []library.Library {
		v, err := xslices.Collect2(libs.All(wnd.Subject()))
		if err != nil {
			alert.ShowBannerError(wnd, err)
			return nil
		}

		return v
	})

	return ui.VStack(
		ui.H2(StrLibraries.Get(wnd)),
		ui.If(!loadedLibs.Valid(), ui.Text(rstring.LabelPleaseWait.Get(wnd))),

		ui.IfFunc(loadedLibs.Valid(), func() core.View {
			return dataview.FromSlice(wnd, loadedLibs.Get(), []dataview.Field[dataview.Element[library.Library]]{
				{
					Name: rstring.LabelName.Get(wnd),
					Map: func(obj dataview.Element[library.Library]) core.View {
						return ui.Text(obj.Value.Name)
					},
				},
			}).NextActionIndicator(true).
				ActionNew(func() {

				}).SelectOptions(
				dataview.NewSelectOptionDelete(wnd, func(selected []dataview.Idx) error {
					for _, i := range selected {
						if idx, ok := i.Int(); ok {
							if err := libs.Delete(wnd.Subject(), loadedLibs.Get()[idx].ID); err != nil {
								return err
							}
						}
					}

					loadedLibs.SetValid(false)

					return nil
				}),
			)
		}),
	).FullWidth().Alignment(ui.Leading)

}
