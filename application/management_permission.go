// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"go.wdy.de/nago/application/permission"
	uipermission "go.wdy.de/nago/application/permission/ui"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/form"
	"iter"
)

type PermissionManagement struct {
	UseCases permission.UseCases
	Pages    uipermission.Pages
}

func (c *Configurator) PermissionManagement() (PermissionManagement, error) {
	if c.permissionManagement == nil {
		c.permissionManagement = &PermissionManagement{
			UseCases: permission.NewUseCases(),
			Pages: uipermission.Pages{
				Permissions: "admin/permissions",
			},
		}

		c.AddContextValue(core.ContextValue("nago.permissions", form.AnyUseCaseList[permission.Permission, permission.ID](func(subject auth.Subject) iter.Seq2[permission.Permission, error] {
			return c.permissionManagement.UseCases.FindAll(subject)
		})))

		c.RootView(c.permissionManagement.Pages.Permissions, c.DecorateRootView(func(wnd core.Window) core.View {
			return uipermission.Permissions(wnd, c.permissionManagement.UseCases.FindAll)
		}))
	}

	return *c.permissionManagement, nil
}
