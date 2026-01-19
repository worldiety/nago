// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiflow

import (
	"slices"

	"go.wdy.de/nago/application/flow"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/colorpicker"
)

func viewTypeStruct(wnd core.Window, uc flow.UseCases, ws *flow.Workspace, m *flow.StructType) core.View {

	addStringFieldPresented := core.AutoState[bool](wnd)
	addBoolFieldPresented := core.AutoState[bool](wnd)
	addTypeFieldPresented := core.AutoState[bool](wnd)
	selectedField := core.AutoState[flow.Field](wnd)

	selectedColor := ui.ColorCardFooter

	return ui.HStack(
		dialogCmd(wnd, ws, "Add String field", addStringFieldPresented, uc.AppendStringField, func() flow.AppendStringFieldCmd {
			return flow.AppendStringFieldCmd{
				Workspace: ws.Identity(),
				Struct:    m.Identity(),
			}
		}),
		dialogCmd(wnd, ws, "Add Bool field", addBoolFieldPresented, uc.AppendBoolField, func() flow.AppendBoolFieldCmd {
			return flow.AppendBoolFieldCmd{
				Workspace: ws.Identity(),
				Struct:    m.Identity(),
			}
		}),

		dialogCmd(wnd, ws, "Add custom type field", addTypeFieldPresented, uc.AppendTypeField, func() flow.AppendTypeFieldCmd {
			return flow.AppendTypeFieldCmd{
				Workspace: ws.Identity(),
				Struct:    m.Identity(),
			}
		}),

		ui.VStack(
			ui.VStack(

				ui.Text(string(m.Name())),
				ui.VStack(
					ui.HStack(
						ui.VStack().BackgroundColor(ui.ColorAccent).Frame(ui.Frame{Width: ui.L8}),
						ui.Grid(
							slices.Collect(func(yield func(cell ui.TGridCell) bool) {
								for field := range m.PrimaryKeyFields() {
									selected := selectedField.Get() != nil && field.Identity() == selectedField.Get().Identity()
									var color ui.Color
									if selected {
										color = selectedColor
									}

									yield(ui.GridCell(hoverWrap(ui.Text(string(field.Name())), func() {
										selectedField.Set(field)
									}).Border(ui.Border{BottomColor: ui.ColorIconsMuted, BottomWidth: ui.L1})).BackgroundColor(color))
									yield(ui.GridCell(viewTypename(field).Border(ui.Border{BottomColor: ui.ColorIconsMuted, BottomWidth: ui.L1})).BackgroundColor(color))
									yield(ui.GridCell(ui.Text(" ").Border(ui.Border{BottomColor: ui.ColorIconsMuted, BottomWidth: ui.L1})).BackgroundColor(color))
								}
							})...,
						).Append(
							slices.Collect(func(yield func(cell ui.TGridCell) bool) {
								for field := range m.NonPrimaryFields() {
									selected := selectedField.Get() != nil && field.Identity() == selectedField.Get().Identity()
									var color ui.Color
									if selected {
										color = selectedColor
									}
									yield(ui.GridCell(hoverWrap(ui.Text(string(field.Name())), func() {
										selectedField.Set(field)
									})).BackgroundColor(color))
									yield(ui.GridCell(viewTypename(field)).BackgroundColor(color))
									if field.Description() != "" {
										yield(ui.GridCell(ui.HStack(ui.ImageIcon(icons.MessageCaption).AccessibilityLabel(field.Description())).TextColor(ui.ColorIconsMuted).Alignment(ui.Trailing)).BackgroundColor(color))
									} else {
										yield(ui.GridCell(ui.Text(" ")).BackgroundColor(color))
									}
								}
							})...,
						).
							Columns(3).
							RowGap(ui.L4).
							FullWidth().
							Padding(ui.Padding{Top: ui.L4, Bottom: ui.L4}),
					).Alignment(ui.Stretch).FullWidth().Gap(ui.L8),
				).
					BackgroundColor(ui.ColorCardBody).
					Border(ui.Border{}.Radius(ui.L8)).
					Padding(ui.Padding{Right: ui.L8}),

				ui.Text(m.Description()).Font(ui.BodySmall).Color(ui.ColorIconsMuted),
			).Alignment(ui.Stretch).Gap(ui.L12),
		).FullWidth(),

		ui.VLine().Frame(ui.Frame{}),
		ui.VStack(
			ui.Text("Struct-Editor"),
			ui.HLine(),
			ui.Menu(
				ui.SecondaryButton(nil).Title("Add field").PreIcon(icons.Plus),
				ui.MenuGroup(
					ui.MenuItem(func() {
						addStringFieldPresented.Set(true)
					}, ui.Text("Add String field")),
					ui.MenuItem(func() {
						addBoolFieldPresented.Set(true)
					}, ui.Text("Add Bool field")),
					ui.MenuItem(func() {
						addTypeFieldPresented.Set(true)
					}, ui.Text("Add Type field")),
				),
			),

			ui.IfFunc(selectedField.Get() != nil, func() core.View {
				return ui.VStack(
					ui.HLine(),
					ui.Text("Field "+string(selectedField.Get().Name())),
					ui.IfFunc(isPrimaryCandidate(selectedField.Get()), func() core.View {
						return ui.SecondaryButton(func() {
							_, err := uc.SelectPrimaryKey(wnd.Subject(), flow.SelectPrimaryKeyCmd{
								Workspace: ws.Identity(),
								Struct:    m.Identity(),
								Field:     selectedField.Get().Identity(),
							})
							if err != nil {
								alert.ShowBannerError(wnd, err)
								return
							}
						}).Title("Set as Primary Key")
					}),
				).FullWidth()
			}),
		).
			Alignment(ui.Top).
			Frame(ui.Frame{Width: ui.L200, MaxWidth: ui.L200}),
	).FullWidth().Alignment(ui.Stretch)

}

func fieldDescView(field flow.Field) core.View {
	if field.Description() != "" {
		return ui.HStack(ui.ImageIcon(icons.MessageCaption).AccessibilityLabel(field.Description())).TextColor(ui.ColorIconsMuted).Alignment(ui.Trailing)
	}
	return ui.Text(" ")
}

func isPrimaryCandidate(field flow.Field) bool {
	_, strType := field.(*flow.StringField)
	return !field.IsPrimaryKey() && strType
}

func viewTypename(field flow.Field) ui.DecoredView {
	var color ui.Color
	switch field.Typename() {
	case "string":
		color = colorpicker.DefaultPalette[0]
	case "bool":
		color = colorpicker.DefaultPalette[1]
	default:
		color = colorpicker.DefaultPalette[2]
	}

	return ui.Text(field.Typename()).Color(color).
		Padding(ui.Padding{Left: ui.L8})
}

func hoverWrap(view core.View, action func()) ui.THStack {
	return ui.HStack(view).Alignment(ui.Leading).HoveredBackgroundColor(ui.I1).Action(action)
}
