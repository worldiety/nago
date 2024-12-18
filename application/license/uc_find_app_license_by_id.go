package license

import (
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/std"
)

func NewFindAppLicenseByID(repo AppLicenseRepository) FindAppLicenseByID {
	return func(subject permission.Auditable, id ID) (std.Option[AppLicense], error) {
		if err := subject.Audit(PermFindAppLicenseByID); err != nil {
			return std.None[AppLicense](), err
		}

		return repo.FindByID(id)
	}
}
