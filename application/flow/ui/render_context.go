// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiflow

import (
	"context"
	"fmt"
	"reflect"

	"go.wdy.de/nago/application/flow"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
)

type RContext struct {
	parent         TFormEditor
	Context        context.Context
	wnd            core.Window
	Handle         flow.HandleCommand
	ws             *flow.Workspace
	selectedStates map[flow.ViewID]*core.State[bool]
	RenderInsert   func(ctx RContext) core.View
	RenderEdit     func(ctx RContext, wrapped core.View) core.View
	RenderDelete   func(ctx RContext, wrapped core.View) core.View
}

func (ctx RContext) Workspace() *flow.Workspace {
	return ctx.ws
}

func (ctx RContext) Form() *flow.Form {
	return ctx.parent.form
}

func (ctx RContext) RenderAppend(parent flow.ViewID) core.View {
	return ui.SecondaryButton(func() {
		ctx.parent.selectedParent.Set(parent)
		ctx.parent.addDialogPresented.Set(true)
	}).Title("Append").PreIcon(icons.Plus)
}

func (ctx RContext) Window() core.Window {
	return ctx.wnd
}

func (ctx RContext) RenderEditor(view flow.FormView) core.View {
	r, ok := ctx.parent.renderersById[reflect.TypeOf(view)]
	if !ok {
		return ui.Text(fmt.Sprintf("%T has no renderer", view))
	}

	return r.Update(ctx, view)
}

func (ctx RContext) RenderPreview(view flow.FormView) core.View {
	r, ok := ctx.parent.renderersById[reflect.TypeOf(view)]
	if !ok {
		return ui.Text(fmt.Sprintf("%T has no renderer", view))
	}

	state := core.StateOf[bool](ctx.wnd, string(view.Identity()))
	wasSelected := state.Get()
	ctx.selectedStates[view.Identity()] = state

	return ui.VStack(
		r.Preview(ctx, view),
	).FullWidth().
		Action(func() {
			for _, c := range ctx.selectedStates {
				c.Set(false)
			}
			state.Set(!wasSelected)
			if state.Get() {
				ctx.parent.selected.Set(view)
			} else {
				ctx.parent.selected.Set(nil)
			}
		}).With(func(stack ui.TVStack) ui.TVStack {
		if state.Get() {
			stack = stack.BackgroundColor(ui.I1)
		}

		return stack
	}).
		Border(ui.Border{}.Radius(ui.L8).
			Color(ui.ColorIconsMuted).
			Width(ui.L1).
			Style(ui.BorderStyleDashed)).
		Padding(ui.Padding{}.All(ui.L16))

}
