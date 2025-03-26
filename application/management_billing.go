// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"go.wdy.de/nago/application/billing"
	uibilling "go.wdy.de/nago/application/billing/ui"
	"go.wdy.de/nago/presentation/core"
)

type BillingManagement struct {
	UseCases billing.UseCases
	Pages    uibilling.Pages
}

func (c *Configurator) BillingManagement() (BillingManagement, error) {
	if c.billingManagement == nil {
		licMgmt, err := c.LicenseManagement()
		if err != nil {
			return BillingManagement{}, err
		}

		usrMgmt, err := c.UserManagement()
		if err != nil {
			return BillingManagement{}, err
		}

		c.billingManagement = &BillingManagement{
			UseCases: billing.NewUseCases(
				usrMgmt.UseCases.SysUser,
				licMgmt.UseCases.PerApp.FindAll,
				licMgmt.UseCases.PerUser.FindAll,
				usrMgmt.UseCases.CountAssignedUserLicense,
			),
			Pages: uibilling.Pages{
				AppLicenses:  "admin/billing/per-app-licenses",
				UserLicenses: "admin/billing/per-user-licenses",
			},
		}

		c.RootViewWithDecoration(c.billingManagement.Pages.AppLicenses, func(wnd core.Window) core.View {
			return uibilling.AppLicensePage(wnd, c.billingManagement.UseCases.AppLicenses)
		})

		c.RootViewWithDecoration(c.billingManagement.Pages.UserLicenses, func(wnd core.Window) core.View {
			return uibilling.UserLicensePage(wnd, c.billingManagement.UseCases.UserLicenses)
		})
	}

	return *c.billingManagement, nil
}
