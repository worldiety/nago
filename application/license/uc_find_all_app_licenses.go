package license

import (
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/xiter"
	"iter"
)

func NewFindAllAppLicenses(repo AppLicenseRepository) FindAllAppLicenses {
	return func(subject permission.Auditable) iter.Seq2[AppLicense, error] {
		if err := subject.Audit(PermFindAllAppLicenses); err != nil {
			return xiter.WithError[AppLicense](err)
		}

		return repo.All()
	}
}
