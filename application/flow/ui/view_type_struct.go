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
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
)

func viewTypeStruct(wnd core.Window, uc flow.UseCases, ws *flow.Workspace, m *flow.StructType) core.View {

	addStringFieldPresented := core.AutoState[bool](wnd)

	return ui.HStack(
		dialogCmd(wnd, ws, "Add String field", addStringFieldPresented, uc.AppendStringField),

		ui.VStack(
			ui.VStack(

				ui.Text(string(m.Name())),
				ui.VStack(
					ui.HStack(
						ui.VStack().BackgroundColor(ui.ColorAccent).Frame(ui.Frame{Width: ui.L8}),
						ui.Grid(
							ui.GridCell(ui.Text("A1").Border(ui.Border{BottomColor: ui.ColorIconsMuted, BottomWidth: ui.L1})),
							ui.GridCell(ui.Text("A2").Border(ui.Border{BottomColor: ui.ColorIconsMuted, BottomWidth: ui.L1})),
							ui.GridCell(ui.Text("B1")),
							ui.GridCell(ui.Text("B2")),
						).Columns(2).FullWidth(),
					).Alignment(ui.Stretch).FullWidth().Gap(ui.L8),
				).
					BackgroundColor(ui.ColorCardBody).
					Border(ui.Border{}.Radius(ui.L8)).
					Padding(ui.Padding{Right: ui.L8}),

				ui.Text(m.Description()).Font(ui.BodySmall).Color(ui.ColorIconsMuted),
			).Alignment(ui.Stretch),
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
				),
			),
		).
			Alignment(ui.Top).
			Frame(ui.Frame{Width: ui.L200, MaxWidth: ui.L200}),
	).FullWidth().Alignment(ui.Stretch)

}
