// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiflow

import (
	"fmt"

	"github.com/worldiety/jsonptr"
	"go.wdy.de/nago/application/flow"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/form"
	"go.wdy.de/nago/presentation/ui/picker"
)

var _ ViewRenderer = (*ButtonRenderer)(nil)

type ButtonRenderer struct {
}

func (r *ButtonRenderer) Create(ctx RContext, parent, after flow.ViewID) (core.View, Apply) {
	f, ok := ctx.Workspace().Forms.ByView(parent)
	if !ok {
		return alert.BannerError(fmt.Errorf("parent has no form: %s", parent)), nil
	}

	state := core.StateOf[flow.AddFormButtonCmd](ctx.Window(), fmt.Sprintf("%s-%s", parent, after)).Init(func() flow.AddFormButtonCmd {
		return flow.AddFormButtonCmd{
			Workspace: ctx.Workspace().Identity(),
			Form:      f.ID,
			Parent:    parent,
			After:     after,
			Style:     "primary",
		}
	})

	errState := core.DerivedState[error](state, "err")

	return ui.VStack(
			form.Auto(form.AutoOptions{
				Errors:  errState.Get(),
				Context: ctx.Context,
			}, state).FullWidth(),
		).FullWidth(), func() error {
			return ctx.Handle(ctx.Window().Subject(), state.Get())
		}
}

func (r *ButtonRenderer) Update(ctx RContext, view flow.FormView) core.View {
	btn := view.(*flow.FormButton)

	alignState := core.StateOf[[]ui.ButtonStyle](ctx.Window(), string(view.Identity())+"style").Init(func() []ui.ButtonStyle {
		return []ui.ButtonStyle{btn.Style()}
	}).Observe(func(newValue []ui.ButtonStyle) {
		var align ui.ButtonStyle
		if len(newValue) > 0 {
			align = newValue[0]
		}

		if err := ctx.HandleCommand(flow.UpdateButtonStyle{
			Workspace: ctx.Workspace().ID,
			Form:      ctx.Form().ID,
			ID:        view.Identity(),
			Style:     align,
		}); err != nil {
			alert.ShowBannerError(ctx.Window(), err)
		}

	})

	return ui.VStack(

		ui.Text("Edit Button"),
		picker.Picker[ui.ButtonStyle]("Style", ui.ButtonStyles(), alignState).Frame(ui.Frame{}.FullWidth()),
	).FullWidth().Gap(ui.L8)
}

func (r *ButtonRenderer) TeaserPreview(ctx RContext) core.View {
	return ui.VStack(
		ui.SecondaryButton(nil).Title("Button"),
	)
}

func (r *ButtonRenderer) Preview(ctx RContext, view flow.FormView) core.View {
	text := view.(*flow.FormButton)
	return ui.Button(text.Style(), nil).Title(text.Title())
}

func (r *ButtonRenderer) Bind(ctx ViewerRenderContext, view flow.FormView, state *core.State[*jsonptr.Obj]) core.View {
	btn := view.(*flow.FormButton)
	return ui.Button(btn.Style(), func() {
		ctx.EvaluateAction(btn)
	}).Title(btn.Title()).
		Enabled(ctx.EvaluateEnabled(btn)).
		Frame(btn.Frame())
}
