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

func editFrameable(ctx RContext, view flow.FormView) core.View {
	frameable, _ := view.(flow.Frameable)
	if frameable == nil {
		return nil
	}

	toggle := core.StateOf[bool](ctx.wnd, "toggle-frame")

	var pushChange func()
	minWidth := core.StateOf[string](ctx.Window(), string(view.Identity())+"minWidth").Init(func() string {
		return string(frameable.Frame().MinWidth)
	}).Observe(func(newValue string) {
		pushChange()
	})

	width := core.StateOf[string](ctx.Window(), string(view.Identity())+"width").Init(func() string {
		return string(frameable.Frame().Width)
	}).Observe(func(newValue string) {
		pushChange()
	})

	maxWidth := core.StateOf[string](ctx.Window(), string(view.Identity())+"maxWidth").Init(func() string {
		return string(frameable.Frame().MaxWidth)
	}).Observe(func(newValue string) {
		pushChange()
	})

	minHeight := core.StateOf[string](ctx.Window(), string(view.Identity())+"minHeight").Init(func() string {
		return string(frameable.Frame().MinHeight)
	}).Observe(func(newValue string) {
		pushChange()
	})

	height := core.StateOf[string](ctx.Window(), string(view.Identity())+"height").Init(func() string {
		return string(frameable.Frame().Height)
	}).Observe(func(newValue string) {
		pushChange()
	})

	maxHeight := core.StateOf[string](ctx.Window(), string(view.Identity())+"maxHeight").Init(func() string {
		return string(frameable.Frame().MaxHeight)
	}).Observe(func(newValue string) {
		pushChange()
	})

	pushChange = func() {
		if err := ctx.HandleCommand(flow.UpdateFormFrame{
			Workspace: ctx.Workspace().ID,
			Form:      ctx.Form().ID,
			ID:        view.Identity(),
			Frame: ui.Frame{
				Width:     ui.Length(width.Get()),
				MinWidth:  ui.Length(minWidth.Get()),
				MaxWidth:  ui.Length(maxWidth.Get()),
				Height:    ui.Length(height.Get()),
				MinHeight: ui.Length(minHeight.Get()),
				MaxHeight: ui.Length(maxHeight.Get()),
			},
		}); err != nil {
			alert.ShowBannerError(ctx.Window(), err)
		}
	}

	return accordion.Accordion(
		ui.Text("Frame"),
		ui.VStack(
			ui.TextField("Width", width.Get()).InputValue(width).FullWidth(),
			ui.TextField("MinWidth", minWidth.Get()).InputValue(minWidth).FullWidth(),
			ui.TextField("MaxWidth", maxWidth.Get()).InputValue(maxWidth).FullWidth(),
			ui.TextField("Height", height.Get()).InputValue(height).FullWidth(),
			ui.TextField("MinHeight", minHeight.Get()).InputValue(minHeight).FullWidth(),
			ui.TextField("MaxHeight", maxHeight.Get()).InputValue(maxHeight).FullWidth(),
		).FullWidth().Alignment(ui.Leading),
		toggle,
	).FullWidth()

}
