// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiflow

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/flow"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/dataview"
	"go.wdy.de/nago/presentation/ui/form"
)

func PageWorkspaces(wnd core.Window, pages Pages, uc flow.UseCases) core.View {
	createPresented := core.AutoState[bool](wnd)

	return ui.VStack(
		createDialog(wnd, createPresented, uc),
		ui.H1(StrWorkspaces.Get(wnd)),
		dataview.FromData(wnd, dataview.Data[*flow.Workspace, flow.WorkspaceID]{
			FindAll: uc.FindWorkspaces(wnd.Subject()),
			FindByID: func(id flow.WorkspaceID) (option.Opt[*flow.Workspace], error) {
				return uc.LoadWorkspace(wnd.Subject(), id)
			},
			Fields: []dataview.Field[*flow.Workspace]{
				{
					ID:   "name",
					Name: rstring.LabelName.Get(wnd),
					Map: func(obj *flow.Workspace) core.View {
						return ui.Text(string(obj.Name))
					},
				},

				{
					ID:   "desc",
					Name: rstring.LabelDescription.Get(wnd),
					Map: func(obj *flow.Workspace) core.View {
						return ui.Text(obj.Description)
					},
				},
			},
		}).CreateAction(func() {
			createPresented.Set(true)
		}).NextActionIndicator(true).Action(func(e *flow.Workspace) {
			wnd.Navigation().ForwardTo(pages.Editor, core.Values{"workspace": string(e.Identity())})
		}).SelectOptions(dataview.NewSelectOptionDelete(wnd, func(selected []flow.WorkspaceID) error {
			for _, id := range selected {
				if err := uc.DeleteWorkspace(wnd.Subject(), id); err != nil {
					return err
				}
			}

			return nil
		})),
	).FullWidth().Alignment(ui.Leading)
}

func createDialog(wnd core.Window, presented *core.State[bool], uc flow.UseCases) core.View {
	if !presented.Get() {
		return nil
	}

	state := core.AutoState[flow.CreateWorkspaceCmd](wnd).Init(func() flow.CreateWorkspaceCmd {
		return flow.CreateWorkspaceCmd{
			ID: data.RandIdent[flow.WorkspaceID](),
		}
	})
	errState := core.DerivedState[error](state, "err")

	return alert.Dialog(
		rstring.ActionCreate.Get(wnd),
		form.Auto(form.AutoOptions{Errors: errState.Get()}, state),
		presented,
		alert.Closeable(),
		alert.Cancel(nil),
		alert.Create(func() (close bool) {
			err := uc.HandleCommand(wnd.Subject(), state.Get())
			errState.Set(err)
			
			if err != nil {
				return false
			}

			return true
		}),
	)
}
