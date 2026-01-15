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
	"go.wdy.de/nago/presentation/ui"
)

func viewTypeString(wnd core.Window, uc flow.UseCases, ws *flow.Workspace, m *flow.StringType) core.View {
	return ui.VStack(
		ui.Text(m.Description()),
		ui.VStack(
			ui.HStack(
				ui.Text(string(m.Name())),
				ui.Space(ui.L20),
				ui.Text("string"),
			),
			ui.HLine(),
		).
			BackgroundColor(ui.ColorCardBody).
			Border(ui.Border{}.Radius(ui.L8)).
			Padding(ui.Padding{}.All(ui.L8)),
	).Alignment(ui.Leading)
}
