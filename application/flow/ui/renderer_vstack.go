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

var (
	_ ViewRenderer = (*VStackRenderer)(nil)
)

type VStackRenderer struct {
}

func (r VStackRenderer) Preview(ctx RContext, view flow.FormView) core.View {
	vstack := view.(*flow.FormVStack)
	var tmp []core.View

	var lastViewID flow.ViewID
	c := 0
	for formView := range vstack.All() {
		if ctx.InsertMode() {
			tmp = append(tmp, ctx.RenderInsertAfter(view.Identity(), lastViewID))
		}

		tmp = append(tmp, ctx.RenderPreview(formView))
		lastViewID = formView.Identity()
		c++
	}

	if ctx.InsertMode() {
		tmp = append(tmp, ctx.RenderInsertAfter(view.Identity(), lastViewID))
	}

	tmp = append(tmp, ctx.SelectedLayer(view.Identity()))

	return ui.VStack(tmp...).
		Position(ui.Position{Type: ui.PositionRelative}).
		Action(ctx.EditorAction(view)).
		Alignment(vstack.Alignment()).
		Gap(vstack.Gap()).
		BackgroundColor(vstack.BackgroundColor()).
		Frame(vstack.Frame()).
		Padding(vstack.Padding()).
		Border(vstack.Border())
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
	_ = view.(*flow.FormVStack)

	return ui.VStack(
		ui.Text("Edit VStack"),
	)
}

func (r VStackRenderer) Bind(ctx ViewerRenderContext, view flow.FormView, state *core.State[*jsonptr.Obj]) core.View {
	vstack := view.(*flow.FormVStack)

	var tmp []core.View
	for formView := range vstack.All() {
		tmp = append(tmp, ctx.Render(formView))
	}

	return ui.VStack(tmp...).
		Alignment(vstack.Alignment()).
		Gap(vstack.Gap()).
		BackgroundColor(vstack.BackgroundColor()).
		Frame(vstack.Frame()).
		Padding(vstack.Padding()).
		Border(vstack.Border())
}
