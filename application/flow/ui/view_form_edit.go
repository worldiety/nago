// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiflow

import (
	"reflect"
	"slices"
	"strings"

	"go.wdy.de/nago/application/flow"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
)

type TFormEditor struct {
	wnd                 core.Window
	uc                  flow.UseCases
	ws                  *flow.Workspace
	form                *flow.Form
	selected            *core.State[flow.ViewID]
	addBelow            *core.State[flow.ViewID]
	addDialogPresented  *core.State[bool]
	selectedRenderer    *core.State[ViewRenderer]
	createViewPresented *core.State[bool]
	selectedParent      *core.State[flow.ViewID]
	selectedAfter       *core.State[flow.ViewID]
	renderersById       map[reflect.Type]ViewRenderer
	renderers           []ViewRenderer
}

func FormEditor(wnd core.Window, opts PageEditorOptions, ws *flow.Workspace, form *flow.Form) TFormEditor {
	c := TFormEditor{
		wnd:                 wnd,
		uc:                  opts.UseCases,
		renderersById:       opts.Renderers,
		form:                form,
		ws:                  ws,
		selected:            core.StateOf[flow.ViewID](wnd, string(ws.Name)+"_nago.flow.form.editor.selected"),
		addBelow:            core.StateOf[flow.ViewID](wnd, string(ws.Name)+"_nago.flow.form.editor.add.below"),
		addDialogPresented:  core.StateOf[bool](wnd, string(ws.Name)+"_nago.flow.form.editor.add.dialog.presented"),
		selectedRenderer:    core.StateOf[ViewRenderer](wnd, string(ws.Name)+"_nago.flow.form.editor.selected.renderer"),
		createViewPresented: core.StateOf[bool](wnd, string(ws.Name)+"_nago.flow.form.editor.create.view.presented"),
		selectedParent:      core.StateOf[flow.ViewID](wnd, string(ws.Name)+"_nago.flow.form.editor.selected.parent"),
		selectedAfter:       core.StateOf[flow.ViewID](wnd, string(ws.Name)+"_nago.flow.form.editor.selected.after"),
	}

	type tmpHolder struct {
		name string
		r    ViewRenderer
	}

	var tmp []tmpHolder

	for t, renderer := range c.renderersById {
		tmp = append(tmp, tmpHolder{
			name: t.String(),
			r:    renderer,
		})
	}

	slices.SortFunc(tmp, func(a, b tmpHolder) int {
		return strings.Compare(a.name, b.name)
	})

	for _, holder := range tmp {
		c.renderers = append(c.renderers, holder.r)
	}

	return c
}

func (c TFormEditor) dialogAddCmd() core.View {
	if !c.createViewPresented.Get() {
		return nil
	}

	view, applyFn := c.selectedRenderer.Get().Create(c.newRenderContext(c.wnd), c.selectedParent.Get(), c.selectedAfter.Get())
	return alert.Dialog(
		"Add form element",
		view,
		c.createViewPresented,
		alert.Closeable(),
		alert.Cancel(func() {
			c.addDialogPresented.Set(true)
		}),
		alert.Create(func() (close bool) {
			if err := applyFn(); err != nil {
				alert.ShowBannerError(c.wnd, err)
				return false
			}

			c.addDialogPresented.Set(false)

			return true
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
			c.selectedRenderer.Set(renderer)
			c.createViewPresented.Set(true)
			c.addDialogPresented.Set(false)

		}).
			HoveredBackgroundColor(ui.I1).
			Border(ui.Border{}.Color(ui.ColorIconsMuted).Width(ui.L1).Radius(ui.L16)).Padding(ui.Padding{}.All(ui.L8)))
	}

	return alert.Dialog(
		"Add form element 2",
		ui.HStack(availableElements...).Wrap(true).Gap(ui.L8).Alignment(ui.Stretch).FullWidth(),
		c.addDialogPresented,
		alert.Larger(),
		alert.Cancel(nil),
		alert.Closeable(),
	)
}

func (c TFormEditor) newRenderContext(wnd core.Window) RContext {
	return RContext{
		parent:         c,
		Context:        wnd.Context(), // TODO merge with dialog_cmd
		wnd:            wnd,
		Handle:         c.uc.HandleCommand,
		ws:             c.ws,
		selectedStates: map[flow.ViewID]*core.State[bool]{},
	}
}

func (c TFormEditor) renderSelectedViewEditor(ctx RContext) core.View {
	deleteFormPresented := core.StateOf[bool](c.wnd, "form_delete_presented")
	if c.selected.Get() == "" {
		return ui.VStack(
			c.deleteFormDialog(deleteFormPresented),
			ui.Heading(6, string(c.form.Name())),
			ui.Text(c.form.Description()),
			ui.HLine(),
			ui.SecondaryButton(func() {
				deleteFormPresented.Set(true)
			}).Title("Delete form"),
		).FullWidth().Alignment(ui.TopLeading)
	}

	deletePresented := core.StateOf[bool](c.wnd, "view_delete_presented")
	conditionalPresented := core.StateOf[bool](c.wnd, "view_conditional_presented")

	view, _ := flow.GetView(c.ws, c.form.ID, c.selected.Get())

	return ui.VStack(
		c.deleteViewDialog(deletePresented),
		c.conditionalFormDialog(conditionalPresented),

		ctx.RenderEditor(view),
		ui.HLine(),
		ui.SecondaryButton(func() {
			conditionalPresented.Set(true)
		}).Title("Conditional Visibility").Frame(ui.Frame{}.FullWidth()),
		ui.SecondaryButton(func() {
			deletePresented.Set(true)
		}).Title("Delete").Frame(ui.Frame{}.FullWidth()),
	).FullWidth().Alignment(ui.TopLeading).Gap(ui.L8)
}

func (c TFormEditor) conditionalFormDialog(presented *core.State[bool]) core.View {
	if !presented.Get() {
		return nil
	}

	code := core.StateOf[string](c.wnd, string(c.selected.Get())+"-visibility-expr-code").Init(func() string {
		view, ok := flow.GetView(c.ws, c.form.ID, c.selected.Get())
		if !ok {
			return ""
		}
		return string(view.VisibleExpr())
	})

	structType, ok := c.ws.Packages.StructTypeByID(c.form.RepositoryType())
	var fieldDetails core.View
	if ok {
		fieldDetails = ui.CodeEditor(structAsGoCode(c.ws, structType)).Language("go").Disabled(true).Frame(ui.Frame{}.FullWidth())
	}

	return alert.Dialog(
		"conditional visibility",
		ui.VStack(
			ui.CodeEditor(code.Get()).InputValue(code).Frame(ui.Frame{}.FullWidth()),
			ui.Text("example: field(\"MyFieldName\").Bool() == true").
				Color(ui.ColorIconsMuted).
				Font(ui.BodySmall),
			ui.Space(ui.L16),
			fieldDetails,
		).FullWidth().Alignment(ui.Leading),
		presented,
		alert.Closeable(),
		alert.Larger(),
		alert.Cancel(nil),
		alert.Save(func() (close bool) {
			if err := c.uc.HandleCommand(c.wnd.Subject(), flow.UpdateFormVisibleExpr{
				Workspace:  c.ws.Identity(),
				Form:       c.form.Identity(),
				ID:         c.selected.Get(),
				Expression: flow.Expression(code.Get()),
			}); err != nil {
				alert.ShowBannerError(c.wnd, err)
			}

			return true
		}),
	)
}

func (c TFormEditor) deleteFormDialog(presented *core.State[bool]) core.View {
	if !presented.Get() {
		return nil
	}

	return alert.Dialog(
		"Delete form",
		ui.Text("Are you sure you want to delete this form?"),
		presented,
		alert.Closeable(),
		alert.Cancel(nil),
		alert.Delete(func() {
			if err := c.uc.HandleCommand(c.wnd.Subject(), flow.DeleteFormCmd{
				Workspace: c.ws.Identity(),
				ID:        c.form.Identity(),
			}); err != nil {
				alert.ShowBannerError(c.wnd, err)
			}
		}),
	)
}

func (c TFormEditor) deleteViewDialog(presented *core.State[bool]) core.View {
	if !presented.Get() {
		return nil
	}

	return alert.Dialog(
		"Delete",
		ui.Text("Are you sure you want to delete this view?"),
		presented,
		alert.Closeable(),
		alert.Cancel(nil),
		alert.Delete(func() {
			if err := c.uc.HandleCommand(c.wnd.Subject(), flow.DeleteViewCmd{
				Workspace: c.ws.Identity(),
				Form:      c.form.Identity(),
				View:      c.selected.Get(),
			}); err != nil {
				alert.ShowBannerError(c.wnd, err)
			}
		}),
	)
}

func (c TFormEditor) Render(ctx core.RenderContext) core.RenderNode {
	rctx := c.newRenderContext(ctx.Window())

	return ui.Grid(
		ui.GridCell(
			ui.VStack(
				c.dialogAddFormElement(),
				c.dialogAddCmd(),
				rctx.RenderPreview(c.form.Root, ui.Center),
			).FullWidth(),
		),
		ui.GridCell(c.renderSelectedViewEditor(rctx)),
	).Widths("1fr", "10rem").
		Columns(2).
		Gap(ui.L8).
		FullWidth().
		Padding(ui.Padding{}.Vertical(ui.L16)).Render(ctx)
}
