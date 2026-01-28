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
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/form"
	"go.wdy.de/nago/presentation/ui/picker"
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
			ctx.RenderPreview(formView, vstack.Alignment()),
		)
		lastViewID = formView.Identity()
	}

	tmp = append(tmp, ctx.RenderInsertAfter(view.Identity(), lastViewID))

	return ui.VStack(tmp...).FullWidth().Gap(ui.L8).Alignment(vstack.Alignment())
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
	vstack := view.(*flow.FormVStack)

	alignState := core.StateOf[[]ui.Alignment](ctx.Window(), string(view.Identity())+"alignment").Init(func() []ui.Alignment {
		return []ui.Alignment{vstack.Alignment()}
	}).Observe(func(newValue []ui.Alignment) {
		var align ui.Alignment
		if len(newValue) > 0 {
			align = newValue[0]
		}

		if err := ctx.HandleCommand(flow.UpdateFormAlignment{
			Workspace: ctx.Workspace().ID,
			Form:      ctx.Form().ID,
			ID:        view.Identity(),
			Alignment: align,
		}); err != nil {
			alert.ShowBannerError(ctx.Window(), err)
		}

	})

	return ui.VStack(
		ui.Text("Edit VStack"),
		picker.Picker[ui.Alignment]("Alignment", ui.Alignments(), alignState),
	)
}

func (r VStackRenderer) Bind(ctx ViewerRenderContext, view flow.FormView, state *core.State[*jsonptr.Obj]) core.View {
	vstack := view.(*flow.FormVStack)

	var tmp []core.View
	for formView := range vstack.All() {
		tmp = append(tmp, ctx.Render(formView))
	}

	return ui.VStack(tmp...).Alignment(vstack.Alignment()).FullWidth()
}
