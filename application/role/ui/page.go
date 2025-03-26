// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uirole

import (
	"go.wdy.de/nago/application/rcrud"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/crud"
)

type Pages struct {
	Roles core.NavigationPath
}

func GroupPage(wnd core.Window, useCases rcrud.UseCases[role.Role, role.ID]) core.View {
	return crud.AutoRootView(crud.AutoRootViewOptions{Title: "Rollen"}, useCases)(wnd)
}
