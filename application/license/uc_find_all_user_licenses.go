package license

import (
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/xiter"
	"iter"
)

func NewFindAllUserLicenses(repo UserLicenseRepository) FindAllUserLicenses {
	return func(subject permission.Auditable) iter.Seq2[UserLicense, error] {
		if err := subject.Audit(PermFindAllUserLicenses); err != nil {
			return xiter.WithError[UserLicense](err)
		}

		return repo.All()
	}
}
