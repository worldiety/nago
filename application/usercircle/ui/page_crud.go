// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiusercircles

import (
	"go.wdy.de/nago/application/rcrud"
	"go.wdy.de/nago/application/usercircle"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/crud"
)

func PageOverview(wnd core.Window, useCases rcrud.UseCases[usercircle.Circle, usercircle.ID]) core.View {
	return crud.AutoRootView(crud.AutoRootViewOptions{Title: "Kreise verwalten"}, useCases)(wnd)
}
