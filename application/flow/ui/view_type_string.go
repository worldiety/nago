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
	"go.wdy.de/nago/presentation/ui/colorpicker"
)

func viewTypeString(wnd core.Window, uc flow.UseCases, ws *flow.Workspace, m *flow.StringType) core.View {
	addStringCasePresented := core.AutoState[bool](wnd)

	return ui.Grid(
		ui.GridCell(
			ui.VStack(
				dialogCmd(wnd, ws, "Add String Case", addStringCasePresented, uc.HandleCommand, func() flow.WorkspaceCommand {
					return flow.AddStringEnumCaseCmd{
						Workspace: ws.Identity(),
						String:    m.Identity(),
					}
				}),

				ui.VStack(

					ui.Text(string(m.Name())),
					ui.VStack(
						ui.HStack(
							ui.VStack().BackgroundColor(ui.ColorAccent).Frame(ui.Frame{Width: ui.L8}),
							ui.Grid(
								slices.Collect(func(yield func(cell ui.TGridCell) bool) {
									yield(ui.GridCell(ui.Text(string(m.Name())).Border(ui.Border{BottomColor: ui.ColorIconsMuted, BottomWidth: ui.L1})))
									yield(ui.GridCell(viewTypename(&flow.StringField{}).Border(ui.Border{BottomColor: ui.ColorIconsMuted, BottomWidth: ui.L1})))
									yield(ui.GridCell(ui.Text(" ").Border(ui.Border{BottomColor: ui.ColorIconsMuted, BottomWidth: ui.L1})))

									for enumCase := range m.Enumeration.All() {
										yield(ui.GridCell(ui.Text(string(enumCase.Name))))
										yield(ui.GridCell(ui.Text(enumCase.Value).Color(colorpicker.DefaultPalette[7]).Padding(ui.Padding{Left: ui.L8})))
										yield(ui.GridCell(ui.Text(enumCase.Description).FullWidth()))
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
			),
		),
		ui.GridCell(ui.VStack(
			ui.Text("Enum"),
			ui.SecondaryButton(func() {
				addStringCasePresented.Set(true)
			}).Title("Add value").PreIcon(icons.Plus),
		)).Alignment(ui.Top),
	).FullWidth().Widths("1fr", "10rem").Columns(2).Gap(ui.L8)

}
