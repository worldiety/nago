// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uilicense

import (
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/application/rcrud"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/crud"
)

type Pages struct {
	AppLicenses  core.NavigationPath
	UserLicenses core.NavigationPath
}

func AppLicensesPage(wnd core.Window, useCases rcrud.UseCases[license.AppLicense, license.ID]) core.View {
	return crud.AutoRootView(crud.AutoRootViewOptions{Title: "Anwendungs-Lizenzen"}, useCases)(wnd)
}

func UserLicensesPage(wnd core.Window, useCases rcrud.UseCases[license.UserLicense, license.ID]) core.View {
	return crud.AutoRootView(crud.AutoRootViewOptions{Title: "Nutzer-Lizenzen"}, useCases)(wnd)
}
