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

func editBorderable(ctx RContext, view flow.FormView) core.View {
	borderable, _ := view.(flow.Borderable)
	if borderable == nil {
		return nil
	}

	toggle := core.StateOf[bool](ctx.wnd, "toggle-borderable")

	var pushChange func()
	allRad := core.StateOf[string](ctx.Window(), string(view.Identity())+"allradius").Init(func() string {
		return string(borderable.Border().TopLeftRadius)
	}).Observe(func(newValue string) {
		pushChange()
	})

	allColor := core.StateOf[string](ctx.Window(), string(view.Identity())+"allcolor").Init(func() string {
		return string(borderable.Border().TopColor)
	}).Observe(func(newValue string) {
		pushChange()
	})

	shadow := core.StateOf[string](ctx.Window(), string(view.Identity())+"shadow").Init(func() string {
		return string(borderable.Border().BoxShadow.Radius)
	})

	pushChange = func() {
		if err := ctx.HandleCommand(flow.UpdateFormBorder{
			Workspace: ctx.Workspace().ID,
			Form:      ctx.Form().ID,
			ID:        view.Identity(),
			Border: ui.Border{}.
				Radius(ui.Length(allRad.Get())).
				Color(ui.Color(allColor.Get())).
				Shadow(ui.Length(shadow.Get())),
		}); err != nil {
			alert.ShowBannerError(ctx.Window(), err)
		}
	}

	return accordion.Accordion(
		ui.Text("Border"),
		ui.VStack(
			ui.TextField("Radius", allRad.Get()).InputValue(allRad).FullWidth(),
			ui.TextField("Color", allColor.Get()).InputValue(allColor).FullWidth(),
			ui.TextField("Shadow", shadow.Get()).InputValue(shadow).FullWidth(),
		).FullWidth().Alignment(ui.Leading),
		toggle,
	).FullWidth()

}
