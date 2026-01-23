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

func viewTypeForm(wnd core.Window, opts PageEditorOptions, ws *flow.Workspace, m *flow.Form) core.View {
	return ui.Grid(
		ui.GridCell(FormEditor(wnd, opts, ws, m)),
		ui.GridCell(ui.Text("right")),
	).Widths("1fr", "10rem").
		Columns(2).
		Gap(ui.L8).
		FullWidth().
		Padding(ui.Padding{}.Vertical(ui.L16))
}
