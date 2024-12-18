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
			UseCases: billing.NewUseCases(usrMgmt.UseCases.SysUser, licMgmt.UseCases.PerApp.FindAll),
			Pages: uibilling.Pages{
				AppLicenses: "admin/billing/per-app-licenses",
			},
		}

		c.RootViewWithDecoration(c.billingManagement.Pages.AppLicenses, func(wnd core.Window) core.View {
			return uibilling.AppLicensePage(wnd, c.billingManagement.UseCases.AppLicenses)
		})
	}

	return *c.billingManagement, nil
}
