package application

import (
	"go.wdy.de/nago/application/license"
	uilicense "go.wdy.de/nago/application/license/ui"
	"go.wdy.de/nago/application/rcrud"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data/json"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"iter"
)

type LicenseManagement struct {
	userLicenseRepo license.UserLicenseRepository
	appLicenseRepo  license.AppLicenseRepository
	UseCases        license.UseCases
	Pages           uilicense.Pages
}

func (c *Configurator) LicenseManagement() (LicenseManagement, error) {
	if c.licenseManagement == nil {
		userLicStore, err := c.EntityStore("nago.iam.license.user")
		if err != nil {
			return LicenseManagement{}, err
		}

		userLicRepo := json.NewSloppyJSONRepository[license.UserLicense, license.ID](userLicStore)

		appLicStore, err := c.EntityStore("nago.iam.license.app")
		if err != nil {
			return LicenseManagement{}, err
		}

		appLicRepo := json.NewSloppyJSONRepository[license.AppLicense, license.ID](appLicStore)

		c.licenseManagement = &LicenseManagement{
			userLicenseRepo: userLicRepo,
			appLicenseRepo:  appLicRepo,
			UseCases:        license.NewUseCases(appLicRepo, userLicRepo),
			Pages: uilicense.Pages{
				AppLicenses:  "admin/iam/license/per-app",
				UserLicenses: "admin/iam/license/per-user",
			},
		}

		c.RootViewWithDecoration(c.licenseManagement.Pages.AppLicenses, func(wnd core.Window) core.View {
			funcs := rcrud.Funcs[license.AppLicense, license.ID]{
				PermFindByID:   license.PermFindAppLicenseByID,
				PermFindAll:    license.PermFindAllAppLicenses,
				PermDeleteByID: license.PermDeleteAppLicense,
				PermCreate:     license.PermCreateAppLicense,
				PermUpdate:     license.PermUpdateAppLicense,
				FindByID: func(subject auth.Subject, id license.ID) (std.Option[license.AppLicense], error) {
					return c.licenseManagement.UseCases.PerApp.FindByID(subject, id)
				},
				FindAll: func(subject auth.Subject) iter.Seq2[license.AppLicense, error] {
					return c.licenseManagement.UseCases.PerApp.FindAll(subject)
				},
				DeleteByID: func(subject auth.Subject, id license.ID) error {
					return c.licenseManagement.UseCases.PerApp.Delete(subject, id)
				},
				Create: func(subject auth.Subject, e license.AppLicense) (license.ID, error) {
					return c.licenseManagement.UseCases.PerApp.Create(subject, e)
				},
				Update: func(subject auth.Subject, e license.AppLicense) error {
					return c.licenseManagement.UseCases.PerApp.Update(subject, e)
				},
				Upsert: func(subject auth.Subject, e license.AppLicense) (license.ID, error) {
					return c.licenseManagement.UseCases.PerApp.Upsert(subject, e)
				},
			}
			return uilicense.AppLicensesPage(wnd, rcrud.UseCasesFrom(&funcs))
		})

		c.RootViewWithDecoration(c.licenseManagement.Pages.UserLicenses, func(wnd core.Window) core.View {
			funcs := rcrud.Funcs[license.UserLicense, license.ID]{
				PermFindByID:   license.PermFindUserLicenseByID,
				PermFindAll:    license.PermFindAllUserLicenses,
				PermDeleteByID: license.PermDeleteUserLicense,
				PermCreate:     license.PermCreateUserLicense,
				PermUpdate:     license.PermUpdateUserLicense,
				FindByID: func(subject auth.Subject, id license.ID) (std.Option[license.UserLicense], error) {
					return c.licenseManagement.UseCases.PerUser.FindByID(subject, id)
				},
				FindAll: func(subject auth.Subject) iter.Seq2[license.UserLicense, error] {
					return c.licenseManagement.UseCases.PerUser.FindAll(subject)
				},
				DeleteByID: func(subject auth.Subject, id license.ID) error {
					return c.licenseManagement.UseCases.PerUser.Delete(subject, id)
				},
				Create: func(subject auth.Subject, e license.UserLicense) (license.ID, error) {
					return c.licenseManagement.UseCases.PerUser.Create(subject, e)
				},
				Update: func(subject auth.Subject, e license.UserLicense) error {
					return c.licenseManagement.UseCases.PerUser.Update(subject, e)
				},
				Upsert: func(subject auth.Subject, e license.UserLicense) (license.ID, error) {
					return c.licenseManagement.UseCases.PerUser.Upsert(subject, e)
				},
			}
			return uilicense.UserLicensesPage(wnd, rcrud.UseCasesFrom(&funcs))
		})
	}

	return *c.licenseManagement, nil
}
