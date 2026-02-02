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

var _ ViewRenderer = (*HStackRenderer)(nil)

type HStackRenderer struct {
}

func (r HStackRenderer) Preview(ctx RContext, view flow.FormView) core.View {
	hstack := view.(*flow.FormHStack)
	var tmp []core.View

	var lastViewID flow.ViewID
	c := 0
	for formView := range hstack.All() {
		tmp = append(tmp,
			ctx.RenderInsertPlus(view.Identity(), lastViewID),
			ctx.RenderPreview(formView, hstack.Alignment()),
		)
		lastViewID = formView.Identity()
		if hstack.Gap() != "" && c < hstack.Len()-1 {
			tmp = append(tmp, ui.VStack(ui.Text(string(hstack.Gap())).Font(ui.BodySmall).Color(ui.ColorIconsMuted)).FullWidth())
		}
		c++
	}

	tmp = append(tmp, ctx.RenderInsertPlus(view.Identity(), lastViewID))

	return ui.HStack(tmp...).FullWidth().Gap(ui.L8).Alignment(hstack.Alignment())
}

func (r HStackRenderer) TeaserPreview(ctx RContext) core.View {
	return ui.VStack(
		ui.Text("Horizontal Stack").Font(ui.BodySmall),
		ui.HStack(
			ui.Text("A"),
			ui.Text("B"),
			ui.Text("C"),
		),
	).FullWidth()
}

func (r HStackRenderer) Create(ctx RContext, parent, after flow.ViewID) (core.View, Apply) {
	state := core.AutoState[flow.AddHStackCmd](ctx.Window()).Init(func() flow.AddHStackCmd {
		return flow.AddHStackCmd{
			Workspace: ctx.Workspace().ID,
			Form:      ctx.Form().ID,
			Parent:    parent,
			After:     after,
		}
	})
	errState := core.DerivedState[error](state, "err")

	return form.Auto[flow.AddHStackCmd](form.AutoOptions{Errors: errState.Get()}, state), func() error {
		err := ctx.parent.uc.HandleCommand(ctx.Window().Subject(), state.Get())
		errState.Set(err)
		return err
	}

}

func (r HStackRenderer) Update(ctx RContext, view flow.FormView) core.View {
	_ = view.(*flow.FormHStack)

	return ui.VStack(
		ui.Text("Edit HStack"),
	)
}

func (r HStackRenderer) Bind(ctx ViewerRenderContext, view flow.FormView, state *core.State[*jsonptr.Obj]) core.View {
	hstack := view.(*flow.FormHStack)

	var tmp []core.View
	for formView := range hstack.All() {
		tmp = append(tmp, ctx.Render(formView))
	}

	return ui.HStack(tmp...).
		Alignment(hstack.Alignment()).
		Gap(hstack.Gap()).
		BackgroundColor(hstack.BackgroundColor()).
		Frame(hstack.Frame()).
		Padding(hstack.Padding()).
		Border(hstack.Border())
}
