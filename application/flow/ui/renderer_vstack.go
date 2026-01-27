// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiflow

import (
	"github.com/worldiety/jsonptr"
	"go.wdy.de/nago/application/flow"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/form"
)

var _ ViewRenderer = (*VStackRenderer)(nil)

type VStackRenderer struct {
}

func (r VStackRenderer) Preview(ctx RContext, view flow.FormView) core.View {
	vstack := view.(*flow.FormVStack)
	var tmp []core.View

	//tmp = append(tmp, ui.Text("VSTACK DEBUG"))
	var lastViewID flow.ViewID
	for formView := range vstack.All() {
		tmp = append(tmp,
			ctx.RenderInsertAfter(view.Identity(), lastViewID),
			ctx.RenderPreview(formView),
		)
		lastViewID = formView.Identity()
	}

	tmp = append(tmp, ctx.RenderInsertAfter(view.Identity(), lastViewID))

	return ui.VStack(tmp...).FullWidth().Gap(ui.L8)
}

func (r VStackRenderer) TeaserPreview(ctx RContext) core.View {
	return ui.VStack(
		ui.Text("Vertical Stack").Font(ui.BodySmall),
		ui.Text("A"),
		ui.Text("B"),
		ui.Text("C"),
	).FullWidth()
}

func (r VStackRenderer) Create(ctx RContext, parent, after flow.ViewID) (core.View, Apply) {
	/*	state := core.DerivedState[flow.WorkspaceCommand](c.createViewPresented, "state").Init(c.addCmdDialogCmd.Get)
		errState := core.DerivedState[error](state, "err")

		return alert.Dialog(
			"Add form element",
			form.Auto(form.AutoOptions{Errors: errState.Get()}, state),
			c.addCmdDialogPresented,
			alert.Closeable(),
			alert.Create(func() (close bool) {
				//var err error
				switch cmd := state.Get().(type) {

				default:
					panic(fmt.Sprintf("cmd %T not implemented", cmd))
				}

				/*			errState.Set(err)
							if err != nil {
								return false
							}

							return true*/
	//}),
	//)*/

	state := core.AutoState[flow.AddVStackCmd](ctx.Window()).Init(func() flow.AddVStackCmd {
		return flow.AddVStackCmd{
			Workspace: ctx.Workspace().ID,
			Form:      ctx.Form().ID,
			Parent:    parent,
			After:     after,
		}
	})
	errState := core.DerivedState[error](state, "err")

	return form.Auto[flow.AddVStackCmd](form.AutoOptions{Errors: errState.Get()}, state), func() error {
		err := ctx.parent.uc.HandleCommand(ctx.Window().Subject(), state.Get())
		errState.Set(err)
		return err
	}

}

func (r VStackRenderer) Update(ctx RContext, view flow.FormView) core.View {
	return ui.Text("Edit VStack")
}

func (r VStackRenderer) Bind(ctx RContext, view flow.ViewID, state *core.State[*jsonptr.Obj]) core.View {
	//TODO implement me
	panic("implement me")
}
