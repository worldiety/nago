// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiflow

import (
	"reflect"

	"github.com/worldiety/jsonptr"
	"go.wdy.de/nago/application/flow"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

type Apply func() error

type TeaserRenderer interface {
	TeaserPreview(ctx RContext) core.View
}

type ViewRenderer interface {
	TeaserRenderer

	Preview(ctx RContext, view flow.FormView) core.View
	Create(ctx RContext, parent, after flow.ViewID) (core.View, Apply)
	Update(ctx RContext, view flow.FormView) core.View
	Bind(ctx RContext, view flow.ViewID, state *core.State[*jsonptr.Obj]) core.View
}

// TODO replace with immutable map
var DefaultRenderers = map[reflect.Type]ViewRenderer{
	reflect.TypeFor[*flow.FormText]():     &TextRenderer{},
	reflect.TypeFor[*flow.FormVStack]():   &VStackRenderer{},
	reflect.TypeFor[*flow.FormCheckbox](): &CheckboxRenderer{},
}

type StringEnumFieldRenderer struct{}

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
