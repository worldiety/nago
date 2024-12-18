package billing

import (
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xiter"
	"iter"
)

func NewAppLicenses(sysUser user.SysUser, licenses license.FindAllAppLicenses) AppLicenses {
	return func(subject auth.Subject) iter.Seq2[license.AppLicense, error] {
		if err := subject.Audit(PermAppLicenses); err != nil {
			return xiter.WithError[license.AppLicense](err)
		}

		return licenses(sysUser())
	}
}
