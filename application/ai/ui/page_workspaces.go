// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiai

import (
	"github.com/worldiety/i18n"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/workspace"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/dataview"
	"go.wdy.de/nago/presentation/ui/form"
	"golang.org/x/text/language"
)

var (
	StrWorkspaces = i18n.MustString("nago.ai.admin.workspaces", i18n.Values{language.English: "Workspaces", language.German: "Workspaces"})
)

type CreateWorkspaceForm struct {
	Name        string             `label:"nago.common.label.name"`
	Description string             `label:"nago.common.label.description" lines:"3"`
	Platform    workspace.Platform `label:"nago.common.label.platform" source:"nago.ai.platforms"`
}

func PageWorkspaces(wnd core.Window, ucWS workspace.UseCases) core.View {
	createDialogPresented := core.AutoState[bool](wnd)

	return ui.VStack(
		ui.H1(StrWorkspaces.Get(wnd)),
		createWorkspaceDialog(wnd, createDialogPresented, ucWS),
		dataview.FromData(wnd, dataview.Data[workspace.Workspace, workspace.ID]{
			FindAll: ucWS.FindAll(wnd.Subject()),
			FindByID: func(id workspace.ID) (option.Opt[workspace.Workspace], error) {
				return ucWS.FindByID(wnd.Subject(), id)
			},
			Fields: []dataview.Field[workspace.Workspace]{
				{
					Name: rstring.LabelName.Get(wnd),
					Map: func(obj workspace.Workspace) core.View {
						return ui.Text(obj.Name)
					},
				},

				{
					Name: rstring.LabelDescription.Get(wnd),
					Map: func(obj workspace.Workspace) core.View {
						return ui.Text(obj.Description)
					},
				},
			},
		}).Selection(true).
			SelectOptions(
				dataview.NewSelectOptionDelete(wnd, func(selected []workspace.ID) error {
					for _, id := range selected {
						if err := ucWS.DeleteByID(wnd.Subject(), id); err != nil {
							return err
						}
					}
					
					return nil
				}),
			).
			ActionNew(func() {
				createDialogPresented.Set(true)
			}),
	).FullWidth().Alignment(ui.Leading)
}

func createWorkspaceDialog(wnd core.Window, presented *core.State[bool], uc workspace.UseCases) core.View {
	if !presented.Get() {
		return nil
	}

	model := core.AutoState[CreateWorkspaceForm](wnd)
	errModel := core.AutoState[error](wnd)

	return alert.Dialog(
		rstring.ActionNew.Get(wnd),
		form.Auto(form.AutoOptions{Window: wnd, Errors: errModel.Get()}, model),
		presented,
		alert.Closeable(),
		alert.Cancel(nil),
		alert.Create(func() (close bool) {
			_, err := uc.Create(wnd.Subject(), workspace.CreateOptions{
				Name:        model.Get().Name,
				Description: model.Get().Description,
				Platform:    model.Get().Platform,
			})

			errModel.Set(err)
			return err == nil
		}),
	)
}
