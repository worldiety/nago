package billing

import (
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"iter"
)

type AppLicenses func(auth.Subject) iter.Seq2[license.AppLicense, error]

type UseCases struct {
	AppLicenses AppLicenses
}

func NewUseCases(sysUser user.SysUser, findAllAppLicences license.FindAllAppLicenses) UseCases {
	return UseCases{
		AppLicenses: NewAppLicenses(sysUser, findAllAppLicences),
	}
}
