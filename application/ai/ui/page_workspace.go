// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiai

import (
	"os"

	"github.com/worldiety/i18n"
	"github.com/worldiety/i18n/date"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/application/ai/workspace"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/breadcrumb"
	"go.wdy.de/nago/presentation/ui/dataview"
	"go.wdy.de/nago/presentation/ui/form"
	"golang.org/x/text/language"
)

var (
	StrWorkspace = i18n.MustString("nago.ai.admin.workspace", i18n.Values{language.English: "Workspace", language.German: "Workspace"})
	StrAgent     = i18n.MustString("nago.ai.admin.agent", i18n.Values{language.English: "Agent", language.German: "Agent"})
	StrAgents    = i18n.MustString("nago.ai.admin.agents", i18n.Values{language.English: "Agents", language.German: "Agents"})
)

type CreateAgentForm struct {
	Name        string `label:"nago.common.label.name"`
	Description string `label:"nago.common.label.description" lines:"3"`
}

func PageWorkspace(wnd core.Window, ucWS workspace.UseCases, ucAgents agent.UseCases) core.View {
	createDialogPresented := core.AutoState[bool](wnd)

	optWS, err := ucWS.FindByID(wnd.Subject(), workspace.ID(wnd.Values()["workspace"]))
	if err != nil {
		return alert.BannerError(err)
	}

	if optWS.IsNone() {
		return alert.BannerError(os.ErrNotExist)
	}

	ws := optWS.Unwrap()

	return ui.VStack(
		ui.H1(StrWorkspace.Get(wnd)),

		breadcrumb.Breadcrumbs(
			ui.TertiaryButton(func() {
				wnd.Navigation().BackwardTo("admin/ai/workspaces", wnd.Values())
			}).Title(StrWorkspaces.Get(wnd)),
		).ClampLeading(),

		ui.Space(ui.L16),

		createAgentDialog(wnd, createDialogPresented, ws, ucWS),
		dataview.FromData(wnd, dataview.Data[agent.Agent, agent.ID]{
			FindAll: xslices.ValuesWithError(ws.Agents, nil),
			FindByID: func(id agent.ID) (option.Opt[agent.Agent], error) {
				return ucAgents.FindByID(wnd.Subject(), id)
			},
			Fields: []dataview.Field[agent.Agent]{
				{
					Name: rstring.LabelName.Get(wnd),
					Map: func(obj agent.Agent) core.View {
						return ui.Text(obj.Name)
					},
				},

				{
					Name: rstring.LabelDescription.Get(wnd),
					Map: func(obj agent.Agent) core.View {
						return ui.Text(obj.Description)
					},
				},
				{
					Name: rstring.LabelState.Get(wnd),
					Map: func(obj agent.Agent) core.View {
						if obj.Error != "" {
							// security note: providing the sync error text may cause an unwanted information disclosure
							return ui.ImageIcon(icons.ExclamationCircle).AccessibilityLabel(obj.Error)
						}

						switch obj.State {
						case agent.StateSynced:
							return ui.ImageIcon(icons.Check).AccessibilityLabel(date.Format(wnd.Locale(), date.Time, obj.LastMod.Time(wnd.Location())))
						default:
							return ui.ImageIcon(icons.CloudArrowUp).AccessibilityLabel(date.Format(wnd.Locale(), date.Time, obj.LastMod.Time(wnd.Location())))
						}

					},
				},
			},
		}).Selection(true).
			SelectOptions(
				dataview.NewSelectOptionDelete(wnd, func(selected []agent.ID) error {
					for _, id := range selected {
						if err := ucWS.DeleteAgent(wnd.Subject(), ws.ID, id); err != nil {
							return err
						}
					}

					return nil
				}),
			).
			Action(func(e agent.Agent) {
				wnd.Navigation().ForwardTo("admin/ai/workspace/agent", wnd.Values().Put("agent", string(e.ID)))
			}).
			ActionNew(func() {
				createDialogPresented.Set(true)
			}),
	).FullWidth().Alignment(ui.Leading)
}

func createAgentDialog(wnd core.Window, presented *core.State[bool], ws workspace.Workspace, uc workspace.UseCases) core.View {
	if !presented.Get() {
		return nil
	}

	model := core.AutoState[CreateAgentForm](wnd)
	errModel := core.AutoState[error](wnd)

	return alert.Dialog(
		rstring.ActionNew.Get(wnd),
		form.Auto(form.AutoOptions{Window: wnd, Errors: errModel.Get()}, model),
		presented,
		alert.Closeable(),
		alert.Cancel(nil),
		alert.Create(func() (close bool) {
			_, err := uc.CreateAgent(wnd.Subject(), ws.ID, workspace.CreateAgentOptions{
				Name:        model.Get().Name,
				Description: model.Get().Description,
				Temperature: 0.7,
			})

			errModel.Set(err)
			return err == nil
		}),
	)
}
