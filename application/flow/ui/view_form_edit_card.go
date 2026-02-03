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

func editCard(ctx RContext, view flow.FormView) core.View {
	paddable, _ := view.(flow.Paddable)
	if paddable == nil {
		return nil
	}
	borderable, _ := view.(flow.Borderable)
	if borderable == nil {
		return nil
	}

	backgroundable, _ := view.(flow.Backgroundable)
	if backgroundable == nil {
		return nil
	}

	state := core.AutoState[bool](ctx.Window())
	toggle := core.StateOf[bool](ctx.wnd, "toggle-card")
	apply := func(bgColor ui.Color, border ui.Border, padding ui.Padding) {
		if err := ctx.HandleCommand(flow.UpdateFormBorder{
			Workspace: ctx.Workspace().ID,
			Form:      ctx.Form().ID,
			ID:        view.Identity(),
			Border:    border,
		}); err != nil {
			alert.ShowBannerError(ctx.Window(), err)
			return
		}

		if err := ctx.HandleCommand(flow.UpdateFormPadding{
			Workspace: ctx.Workspace().ID,
			Form:      ctx.Form().ID,
			ID:        view.Identity(),
			Padding:   padding,
		}); err != nil {
			alert.ShowBannerError(ctx.Window(), err)
			return
		}

		if err := ctx.HandleCommand(flow.UpdateFormBackgroundColorCmd{
			Workspace: ctx.Workspace().ID,
			Form:      ctx.Form().ID,
			ID:        view.Identity(),
			Color:     bgColor,
		}); err != nil {
			alert.ShowBannerError(ctx.Window(), err)
			return
		}

		state.Invalidate()
	}

	return accordion.Accordion(
		ui.Text("Card Style Presets"),
		ui.VStack(
			ui.SecondaryButton(func() {
				apply(
					ui.M4,
					ui.Border{}.Radius(ui.L8),
					ui.Padding{}.All(ui.L16),
				)
			}).Title("Default card").FullWidth(),

			ui.SecondaryButton(func() {
				apply(
					"",
					ui.Border{},
					ui.Padding{},
				)
			}).Title("Clear").FullWidth(),
		).FullWidth().Alignment(ui.Leading).Gap(ui.L8),
		toggle,
	).FullWidth()

}
