package billing

import (
	"fmt"
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"slices"
	"strings"
)

func NewUserLicenses(sysUser user.SysUser, licenses license.FindAllUserLicenses, count user.CountAssignedUserLicense) UserLicenses {
	return func(subject auth.Subject) (UserLicenseStatistics, error) {
		if err := subject.Audit(PermUserLicenses); err != nil {
			return UserLicenseStatistics{}, err
		}

		var res UserLicenseStatistics
		for userLicense, err := range licenses(sysUser()) {
			if err != nil {
				return UserLicenseStatistics{}, fmt.Errorf("cannot get all user licenses: %w", err)
			}

			used, err := count(sysUser(), userLicense.ID)
			if err != nil {
				return UserLicenseStatistics{}, fmt.Errorf("cannot get user license count: %w", err)
			}

			res.Stats = append(res.Stats, PerUserLicenseStats{
				License: userLicense,
				Used:    used,
			})
		}

		slices.SortFunc(res.Stats, func(e PerUserLicenseStats, e2 PerUserLicenseStats) int {
			return strings.Compare(e.License.Name, e2.License.Name)
		})

		return res, nil
	}
}
