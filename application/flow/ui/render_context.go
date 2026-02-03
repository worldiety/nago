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
	insertMode     bool
	inspectorMode  bool
}

func (ctx RContext) InsertMode() bool {
	return ctx.insertMode
}

func (ctx RContext) InspectorMode() bool {
	return ctx.inspectorMode
}

func (ctx RContext) Workspace() *flow.Workspace {
	return ctx.ws
}

func (ctx RContext) Form() *flow.Form {
	return ctx.parent.form
}

func (ctx RContext) RenderInsertAfter(parent flow.ViewID, after flow.ViewID) core.View {
	return ui.HStack(
		/*ui.HLine().Border(ui.Border{TopWidth: "1px", TopColor: ui.ColorInteractive}),
		ui.SecondaryButton(func() {
			ctx.parent.selectedAfter.Set(after)
			ctx.parent.selectedParent.Set(parent)
			ctx.parent.addBelow.Set(after)
			ctx.parent.addDialogPresented.Set(true)
		}).PreIcon(icons.Plus),*/
		ui.SecondaryButton(func() {
			ctx.parent.selectedAfter.Set(after)
			ctx.parent.selectedParent.Set(parent)
			ctx.parent.addBelow.Set(after)
			ctx.parent.addDialogPresented.Set(true)
		}).PreIcon(icons.Plus).Frame(ui.Frame{MinWidth: ui.L40}),
	).FullWidth()
}

func (ctx RContext) RenderInsertPlus(parent flow.ViewID, after flow.ViewID) core.View {
	return ui.SecondaryButton(func() {
		ctx.parent.selectedAfter.Set(after)
		ctx.parent.selectedParent.Set(parent)
		ctx.parent.addBelow.Set(after)
		ctx.parent.addDialogPresented.Set(true)
	}).PreIcon(icons.Plus).Frame(ui.Frame{MinWidth: ui.L40})
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

func (ctx RContext) EditorAction(view flow.FormView) func() {
	state := core.StateOf[bool](ctx.wnd, string(view.Identity()))
	wasSelected := state.Get()
	ctx.selectedStates[view.Identity()] = state

	return func() {
		for _, c := range ctx.selectedStates {
			c.Set(false)
		}
		state.Set(!wasSelected)
		if state.Get() {
			ctx.parent.selected.Set(view.Identity())
		} else {
			ctx.parent.selected.Set("")
		}
	}
}

func (ctx RContext) RenderPreview(view flow.FormView) core.View {
	r, ok := ctx.parent.renderersById[reflect.TypeOf(view)]
	if !ok {
		return ui.Text(fmt.Sprintf("%T has no renderer", view))
	}

	if !ctx.InspectorMode() {
		return r.Preview(ctx, view)
	}

	state := core.StateOf[bool](ctx.wnd, string(view.Identity()))

	return ui.VStack(
		r.Preview(ctx, view),
	).
		Action(ctx.EditorAction(view)).With(func(stack ui.TVStack) ui.TVStack {
		if state.Get() {
			stack = stack.BackgroundColor(ui.I1)
		}

		return stack
	}).FullWidth().Border(ui.Border{}.Radius(ui.L8).
		Color(ui.ColorIconsMuted).
		Width(ui.L1).
		Style(ui.BorderStyleDashed)).
		Padding(ui.Padding{}.All(ui.L16))
}

func (ctx RContext) HandleCommand(cmd flow.WorkspaceCommand) error {
	return ctx.parent.uc.HandleCommand(ctx.wnd.Subject(), cmd)
}
