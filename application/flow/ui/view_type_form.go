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
)

func viewTypeForm(wnd core.Window, opts PageEditorOptions, ws *flow.Workspace, m *flow.Form) core.View {
	return FormEditor(wnd, opts, ws, m)
}
