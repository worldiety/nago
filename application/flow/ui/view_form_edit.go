// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiflow

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/worldiety/jsonptr"
	"go.wdy.de/nago/application/flow"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/picker"
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
	formEditorFlagInsertModeState := core.StateOf[bool](wnd, formEditorFlagInsertMode)
	formEditorFlagShowInspectorState := core.StateOf[bool](wnd, formEditorFlagShowInspector)

	var rinc int
	return RContext{
		parent:           c,
		Context:          wnd.Context(), // TODO merge with dialog_cmd
		wnd:              wnd,
		Handle:           c.uc.HandleCommand,
		ws:               c.ws,
		selectedStates:   map[flow.ViewID]*core.State[bool]{},
		insertMode:       formEditorFlagInsertModeState.Get(),
		inspectorMode:    formEditorFlagShowInspectorState.Get(),
		recursionCounter: &rinc,
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

	view, _ := flow.GetView(c.ws, c.form.ID, c.selected.Get())

	if view == nil {
		return ui.Text("State has no view")
	}

	deletePresented := core.StateOf[bool](c.wnd, "view_delete_presented")
	conditionalPresented := core.StateOf[bool](c.wnd, "view_conditional_presented")
	actionPresented := core.StateOf[bool](ctx.Window(), string(view.Identity())+"action-presented")
	enabledPresented := core.StateOf[bool](ctx.Window(), string(view.Identity())+"enabled-presented")

	alignable, _ := view.(flow.Alignable)
	actionable, _ := view.(flow.Actionable)
	enabler, _ := view.(flow.Enabler)
	gapable, _ := view.(flow.Gapable)
	backgroundable, _ := view.(flow.Backgroundable)

	gapState := core.StateOf[string](ctx.Window(), string(view.Identity())+"gap").Init(func() string {
		if gapable != nil {
			return string(gapable.Gap())
		}

		return ""
	}).Observe(func(newValue string) {
		if gapable == nil {
			return
		}

		if gapable.Gap() == ui.Length(newValue) {
			return
		}

		if err := ctx.HandleCommand(flow.UpdateFormGap{
			Workspace: ctx.Workspace().ID,
			Form:      ctx.Form().ID,
			ID:        view.Identity(),
			Gap:       ui.Length(newValue),
		}); err != nil {
			alert.ShowBannerError(ctx.Window(), err)
		}
	})

	alignState := core.StateOf[[]ui.Alignment](ctx.Window(), string(view.Identity())+"alignment").Init(func() []ui.Alignment {
		if alignable == nil {
			return nil
		}

		return []ui.Alignment{alignable.Alignment()}
	}).Observe(func(newValue []ui.Alignment) {
		if alignable == nil {
			return
		}

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

	hasLayout := gapable != nil || alignable != nil

	backgroundColorState := core.StateOf[string](ctx.Window(), string(view.Identity())+"background-color").Init(func() string {
		if backgroundable == nil {
			return ""
		}

		return string(backgroundable.BackgroundColor())
	}).Observe(func(newValue string) {
		if backgroundable == nil {
			return
		}

		if err := ctx.HandleCommand(flow.UpdateFormBackgroundColorCmd{
			Workspace: ctx.Workspace().ID,
			Form:      ctx.Form().ID,
			ID:        view.Identity(),
			Color:     ui.Color(newValue),
		}); err != nil {
			alert.ShowBannerError(ctx.Window(), err)
		}
	})

	return ui.VStack(
		c.deleteViewDialog(deletePresented),
		c.conditionalFormDialog(conditionalPresented),
		c.formActionDialog(ctx, actionable, actionPresented),
		c.enabledFormDialog(enabledPresented),

		ctx.RenderEditor(view),

		ui.If(hasLayout, ui.HLine()),
		ui.If(hasLayout, ui.Text("Layout")),
		ui.If(alignable != nil,
			picker.Picker[ui.Alignment]("Alignment", ui.Alignments(), alignState).FullWidth(),
		),
		ui.If(gapable != nil, ui.TextField("Gap", gapState.Get()).InputValue(gapState)),

		ui.IfFunc(backgroundable != nil, func() core.View {
			return ui.VStack(
				ui.HLine(),
				ui.Text("Background"),
				ui.TextField("Background Color", backgroundColorState.Get()).InputValue(backgroundColorState).FullWidth(),
			).FullWidth().Alignment(ui.Leading)
		}),

		editBorderable(ctx, view),
		editFrameable(ctx, view),
		editPaddable(ctx, view),
		editCard(ctx, view),

		ui.HLine(),
		ui.Text("Scripts"),
		ui.If(enabler != nil, ui.SecondaryButton(func() {
			enabledPresented.Set(true)
		}).Title("Enabled").FullWidth()),
		ui.If(actionable != nil, ui.SecondaryButton(func() {
			actionPresented.Set(true)
		}).Title("Action").FullWidth()),
		ui.SecondaryButton(func() {
			conditionalPresented.Set(true)
		}).Title("Visibility").FullWidth(),
		ui.HLine(),
		ui.SecondaryButton(func() {
			deletePresented.Set(true)
		}).Title("Delete").FullWidth(),
	).FullWidth().
		Alignment(ui.TopLeading).
		Gap(ui.L8)
}

func (c TFormEditor) Render(ctx core.RenderContext) core.RenderNode {
	formEditorFlagTestState := core.StateOf[bool](ctx.Window(), formEditorFlagTest)
	if formEditorFlagTestState.Get() {

		structType, ok := c.ws.Packages.StructTypeByID(c.form.Type())
		if !ok {
			return alert.BannerError(fmt.Errorf("form referenced a struct type over repository which cannot be resolved: %s.%s", c.form.Type(), c.form.ID)).Render(ctx)
		}

		tmpState := core.StateOf[*jsonptr.Obj](ctx.Window(), string(structType.ID)+"tmp").Init(func() *jsonptr.Obj {
			return jsonptr.NewObj(make(map[string]jsonptr.Value))
		})

		rctx := NewViewerRenderContext(
			ctx.Window().Context(),
			ctx.Window(),
			c.ws,
			c.form,
			structType,
			c.renderersById,
			false,
			tmpState,
		)

		return ui.HStack(
			rctx.Render(c.form.Root),
			ui.VStack(
				ui.CodeEditor(tmpState.Get().String()).Disabled(true).Language("json").FullWidth(),
			).
				BackgroundColor(ui.M3).
				Position(ui.Position{Type: ui.PositionFixed, Right: "1rem", Top: "10rem"}).
				Frame(ui.Frame{Width: ui.L400, MaxWidth: ui.L400}).
				Padding(ui.Padding{}.All(ui.L8)).
				Border(ui.Border{}.Radius(ui.L16).Shadow(ui.L8)),
		).FullWidth().Render(ctx)
	}

	rctx := c.newRenderContext(ctx.Window())

	return ui.Grid(
		ui.GridCell(
			ui.VStack(
				c.dialogAddFormElement(),
				c.dialogAddCmd(),
				rctx.RenderPreview(c.form.Root),
			).FullWidth(),
		),
		ui.GridCell(
			ui.VStack(
				ui.ScrollView(
					c.renderSelectedViewEditor(rctx),
				).Frame(ui.Frame{}.FullWidth()),
			).Alignment(ui.TopLeading).
				BackgroundColor(ui.M3).
				Position(ui.Position{Type: ui.PositionFixed, Right: "1rem", Top: "10rem", Bottom: "1rem"}).
				Frame(ui.Frame{Width: ui.L256, MaxWidth: ui.L256}).
				Padding(ui.Padding{}.All(ui.L8)).
				Border(ui.Border{}.Radius(ui.L16).Shadow(ui.L8)),
		),
	).Widths("1fr", "18rem").
		Columns(2).
		Gap(ui.L8).
		FullWidth().
		Padding(ui.Padding{}.Vertical(ui.L16)).Render(ctx)
}
