// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiflow

import (
	"go.wdy.de/nago/application/flow"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/accordion"
	"go.wdy.de/nago/presentation/ui/alert"
)

func editPaddable(ctx RContext, view flow.FormView) core.View {
	paddable, _ := view.(flow.Paddable)
	if paddable == nil {
		return nil
	}

	toggle := core.StateOf[bool](ctx.wnd, "toggle-pad")

	var pushChange func()
	padLeft := core.StateOf[string](ctx.Window(), string(view.Identity())+"padLeft").Init(func() string {
		return string(paddable.Padding().Left)
	}).Observe(func(newValue string) {
		pushChange()
	})

	padTop := core.StateOf[string](ctx.Window(), string(view.Identity())+"padTop").Init(func() string {
		return string(paddable.Padding().Top)
	}).Observe(func(newValue string) {
		pushChange()
	})

	padRight := core.StateOf[string](ctx.Window(), string(view.Identity())+"padRight").Init(func() string {
		return string(paddable.Padding().Right)
	}).Observe(func(newValue string) {
		pushChange()
	})

	padBot := core.StateOf[string](ctx.Window(), string(view.Identity())+"padBot").Init(func() string {
		return string(paddable.Padding().Bottom)
	}).Observe(func(newValue string) {
		pushChange()
	})

	pushChange = func() {
		if err := ctx.HandleCommand(flow.UpdateFormPadding{
			Workspace: ctx.Workspace().ID,
			Form:      ctx.Form().ID,
			ID:        view.Identity(),
			Padding: ui.Padding{
				Left:   ui.Length(padLeft.Get()),
				Top:    ui.Length(padTop.Get()),
				Right:  ui.Length(padRight.Get()),
				Bottom: ui.Length(padBot.Get()),
			},
		}); err != nil {
			alert.ShowBannerError(ctx.Window(), err)
		}
	}

	return accordion.Accordion(
		ui.Text("Padding"),
		ui.VStack(
			ui.TextField("Left", padLeft.Get()).InputValue(padLeft).FullWidth(),
			ui.TextField("Top", padTop.Get()).InputValue(padTop).FullWidth(),
			ui.TextField("Right", padRight.Get()).InputValue(padRight).FullWidth(),
			ui.TextField("Bottom", padBot.Get()).InputValue(padBot).FullWidth(),
		).FullWidth().Alignment(ui.Leading),
		toggle,
	).FullWidth()

}
