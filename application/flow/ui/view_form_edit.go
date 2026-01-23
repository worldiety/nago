// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiflow

import (
	"fmt"
	"slices"
	"strings"

	"go.wdy.de/nago/application/flow"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/form"
)

type TFormEditor struct {
	wnd                   core.Window
	uc                    flow.UseCases
	ws                    *flow.Workspace
	form                  *flow.Form
	selected              *core.State[flow.FormView]
	addBelow              *core.State[flow.ViewID]
	addDialogPresented    *core.State[bool]
	addCmdDialogPresented *core.State[bool]
	addCmdDialogCmd       *core.State[flow.WorkspaceCommand]
	renderersById         map[flow.RendererID]ViewRenderer
	renderers             []ViewRenderer
}

func FormEditor(wnd core.Window, opts PageEditorOptions, ws *flow.Workspace, form *flow.Form) TFormEditor {
	c := TFormEditor{
		wnd:                   wnd,
		uc:                    opts.UseCases,
		renderersById:         opts.Renderers,
		form:                  form,
		ws:                    ws,
		selected:              core.StateOf[flow.FormView](wnd, string(ws.Name)+"_nago.flow.form.editor.selected"),
		addBelow:              core.StateOf[flow.ViewID](wnd, string(ws.Name)+"_nago.flow.form.editor.add.below"),
		addDialogPresented:    core.StateOf[bool](wnd, string(ws.Name)+"_nago.flow.form.editor.add.dialog.presented"),
		addCmdDialogPresented: core.StateOf[bool](wnd, string(ws.Name)+"_nago.flow.form.editor.add.cmd.dialog.presented"),
		addCmdDialogCmd:       core.StateOf[flow.WorkspaceCommand](wnd, string(ws.Name)+"_nago.flow.form.editor.add.cmd.dialog.cmd"),
	}

	for _, renderer := range c.renderersById {
		c.renderers = append(c.renderers, renderer)
	}

	slices.SortFunc(c.renderers, func(a, b ViewRenderer) int {
		return strings.Compare(string(a.Identity()), string(b.Identity()))
	})

	return c
}

func (c TFormEditor) renderElement(ctx RContext, elem flow.FormView) core.View {
	r, ok := c.renderersById[elem.Renderer()]
	if !ok {
		return ui.Text(fmt.Sprintf("%T refers to unknown renderer '%s'", elem, elem.Renderer()))
	}

	return r.Preview(ctx, elem)
}

func (c TFormEditor) insertViewLarge(elem flow.FormView) core.View {
	return ui.HStack(
		ui.ImageIcon(icons.Plus),
		ui.Text("Add form element"),
	).FullWidth().
		Action(func() {
			c.addBelow.Set("")
			c.selected.Set(elem)
			c.addDialogPresented.Set(true)
		}).
		HoveredBorder(ui.Border{}.Color(ui.I0).Width(ui.L1).Radius(ui.L16)).
		HoveredBackgroundColor(ui.I1).
		Padding(ui.Padding{}.All(ui.L16)).
		Border(ui.Border{}.Color(ui.ColorIconsMuted).Width(ui.L1).Radius(ui.L16))
}

func insertView() core.View {
	return ui.HStack(
		ui.HLine().Border(ui.Border{TopWidth: "1px", TopColor: ui.ColorInteractive}),
		ui.SecondaryButton(func() {

		}).PreIcon(icons.Plus),
	).FullWidth()
}

func (c TFormEditor) dialogAddCmd() core.View {
	if !c.addCmdDialogPresented.Get() {
		return nil
	}

	state := core.DerivedState[flow.WorkspaceCommand](c.addCmdDialogPresented, "state").Init(c.addCmdDialogCmd.Get)
	errState := core.DerivedState[error](state, "err")

	return alert.Dialog(
		"Add form element",
		form.Auto(form.AutoOptions{Errors: errState.Get()}, state),
		c.addCmdDialogPresented,
		alert.Closeable(),
		alert.Create(func() (close bool) {
			//var err error
			switch cmd := state.Get().(type) {

			default:
				panic(fmt.Sprintf("cmd %T not implemented", cmd))
			}

			/*			errState.Set(err)
						if err != nil {
							return false
						}

						return true*/
		}),
	)

}

func (c TFormEditor) dialogAddFormElement() core.View {
	if !c.addDialogPresented.Get() {
		return nil
	}

	rctx := c.newRenderContext(c.wnd)
	var availableElements []core.View
	for _, renderer := range c.renderers {
		availableElements = append(availableElements, ui.VStack(
			renderer.TeaserPreview(rctx),
		).Action(func() {
			//cmd := renderer.CreateCmd(c.ws, c.form.Identity(), c.selected.Get().Identity(), c.addBelow.Get())
			//c.addCmdDialogCmd.Set(cmd)
			c.addCmdDialogPresented.Set(true)
		}).
			HoveredBackgroundColor(ui.I1).
			Border(ui.Border{}.Color(ui.ColorIconsMuted).Width(ui.L1).Radius(ui.L16)).Padding(ui.Padding{}.All(ui.L8)))
	}

	return alert.Dialog(
		"Add form element",
		ui.HStack(availableElements...).Wrap(true).Gap(ui.L8).Alignment(ui.Stretch).FullWidth(),
		c.addDialogPresented,
		alert.Larger(),
		alert.Cancel(nil),
		alert.Closeable(),
	)
}

func (c TFormEditor) newRenderContext(wnd core.Window) RContext {
	return RContext{
		Context:   wnd.Context(), // TODO merge with dialog_cmd
		Window:    wnd,
		Handle:    c.uc.HandleCommand,
		Workspace: c.ws,
	}
}

func (c TFormEditor) Render(ctx core.RenderContext) core.RenderNode {
	rctx := c.newRenderContext(ctx.Window())

	return ui.VStack(
		c.dialogAddFormElement(),
		c.dialogAddCmd(),
		c.renderElement(rctx, c.form.Root),
	).FullWidth().Render(ctx)
}
