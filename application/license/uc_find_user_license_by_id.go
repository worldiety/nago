package license

import (
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/std"
)

func NewFindUserLicenseByID(repo UserLicenseRepository) FindUserLicenseByID {
	return func(subject permission.Auditable, id ID) (std.Option[UserLicense], error) {
		if err := subject.Audit(PermFindUserLicenseByID); err != nil {
			return std.None[UserLicense](), err
		}

		return repo.FindByID(id)
	}
}
