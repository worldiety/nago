// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiflow

import (
	"context"
	"slices"

	"github.com/worldiety/jsonptr"
	"go.wdy.de/nago/application/flow"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

type RContext struct {
	Context   context.Context
	Window    core.Window
	Handle    flow.HandleCommand
	Workspace *flow.Workspace
}

type Apply func() error

type ViewRenderer interface {
	Identity() flow.RendererID
	TeaserPreview(ctx RContext) core.View
	Preview(ctx RContext, id flow.FormView) core.View
	Create(ctx RContext, parent, after flow.ViewID) (core.View, Apply)
	Update(ctx RContext, view flow.ViewID) (core.View, Apply)
	Bind(ctx RContext, view flow.ViewID, state *core.State[*jsonptr.Obj]) core.View
}

var DefaultRenderers = slices.Values([]ViewRenderer{
	&TextRenderer{},
})

type StringEnumFieldRenderer struct{}

func (r StringEnumFieldRenderer) Identity() flow.RendererID {
	return "nago.flow.ui.renderer.string-enum"
}

func (r StringEnumFieldRenderer) CreateCmd(ws *flow.Workspace, form flow.FormID, parent, after flow.FormView) flow.WorkspaceCommand {
	//TODO implement me
	panic("implement me")
}

func (r StringEnumFieldRenderer) Preview(wnd core.Window, view flow.FormView) core.View {
	return ui.Text("Enum String Field").Font(ui.BodySmall)
}

func (r StringEnumFieldRenderer) TeaserPreview(wnd core.Window) core.View {
	return ui.VStack(
		ui.Text("Enum String Field").Font(ui.BodySmall),
		ui.HStack(
			ui.VStack(ui.Text("Option A")).
				Padding(ui.Padding{}.All(ui.L8)).
				Border(ui.Border{}.Radius(ui.L8).Width(ui.L1)),

			ui.VStack(ui.Text("Option B")).
				Padding(ui.Padding{}.All(ui.L8)).
				Border(ui.Border{}.Radius(ui.L8).Width(ui.L1)),
		).FullWidth().Gap(ui.L8),
	)
}

func (r StringEnumFieldRenderer) RenderEdit(wnd core.Window, elem flow.FormView) core.View {
	/*strField := elem.(*flow.TypeField)
	enum := strField.Type().(*flow.StringType)

	var views []core.View
	for enumCase := range enum.Values() {
		views = append(views, ui.VStack(
			ui.Text(enumCase.Description()),
		))
	}

	return ui.VStack(
		ui.Text("Enum String Field").Font(ui.BodySmall),
		ui.HStack(views...).Wrap(true).FullWidth().Gap(ui.L8),
	)*/

	panic("implement me")
}

func (r StringEnumFieldRenderer) CanRender(field flow.Field) bool {

	return false
}

type VStackRenderer struct {
}

func (r VStackRenderer) Preview(wnd core.Window) core.View {
	return ui.VStack(
		ui.Text("A"),
		ui.Text("B"),
		ui.Text("C"),
	).FullWidth()
}

func (r VStackRenderer) CreateCmd(ws *flow.Workspace, form flow.FormID, parent, after flow.ViewID) flow.WorkspaceCommand {
	panic("implement me")
}

func (r VStackRenderer) RenderEdit(wnd core.Window, elem flow.FormView) core.View {
	return nil
}

func (r VStackRenderer) CanRender(field flow.Field) bool {
	//TODO implement me
	panic("implement me")
}

func (r VStackRenderer) Identity() flow.RendererID {
	return "nago.flow.ui.renderer.vstack"
}
