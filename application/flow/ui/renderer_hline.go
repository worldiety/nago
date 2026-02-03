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

var _ ViewRenderer = (*HLineRenderer)(nil)

type HLineRenderer struct {
}

func (r HLineRenderer) Preview(ctx RContext, view flow.FormView) core.View {
	hline := view.(*flow.FormHLine)

	return ui.VStack(ui.TDivider{}.
		Border(ui.Border{TopWidth: "1px", TopColor: ui.M5}).
		Frame(hline.Frame()).
		Padding(hline.Padding())).FullWidth().Action(ctx.EditorAction(view))
}

func (r HLineRenderer) TeaserPreview(ctx RContext) core.View {
	return ui.VStack(
		ui.Text("Horizontal Line").Font(ui.BodySmall),
		ui.HLine(),
	).FullWidth()
}

func (r HLineRenderer) Create(ctx RContext, parent, after flow.ViewID) (core.View, Apply) {
	state := core.AutoState[flow.AddHLineCmd](ctx.Window()).Init(func() flow.AddHLineCmd {
		return flow.AddHLineCmd{
			Workspace: ctx.Workspace().ID,
			Form:      ctx.Form().ID,
			Parent:    parent,
			After:     after,
		}
	})
	errState := core.DerivedState[error](state, "err")

	return form.Auto[flow.AddHLineCmd](form.AutoOptions{Errors: errState.Get()}, state), func() error {
		err := ctx.parent.uc.HandleCommand(ctx.Window().Subject(), state.Get())
		errState.Set(err)
		return err
	}

}

func (r HLineRenderer) Update(ctx RContext, view flow.FormView) core.View {
	_ = view.(*flow.FormHLine)

	return ui.VStack(
		ui.Text("Edit HLine"),
	)
}

func (r HLineRenderer) Bind(ctx ViewerRenderContext, view flow.FormView, state *core.State[*jsonptr.Obj]) core.View {
	hline := view.(*flow.FormHLine)

	return ui.TDivider{}.
		Border(ui.Border{TopWidth: "1px", TopColor: ui.M5}).
		Frame(hline.Frame()).
		Padding(hline.Padding())
}
