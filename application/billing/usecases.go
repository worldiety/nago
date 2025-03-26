// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package billing

import (
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"iter"
)

type AppLicenses func(auth.Subject) iter.Seq2[license.AppLicense, error]
type UserLicenses func(subject auth.Subject) (UserLicenseStatistics, error)

type UserLicenseStatistics struct {
	Stats []PerUserLicenseStats
}

type PerUserLicenseStats struct {
	License license.UserLicense
	Used    int
}

func (p PerUserLicenseStats) Depleted() bool {
	return p.Used == p.License.MaxUsers
}

func (p PerUserLicenseStats) Overcommitted() bool {
	return p.Used > p.License.MaxUsers
}

func (p PerUserLicenseStats) Avail() int {
	return p.License.MaxUsers - p.Used
}

type UseCases struct {
	AppLicenses  AppLicenses
	UserLicenses UserLicenses
}

func NewUseCases(
	sysUser user.SysUser,
	findAllAppLicences license.FindAllAppLicenses,
	findAllUserLicences license.FindAllUserLicenses,
	countAssignedUserLicense user.CountAssignedUserLicense,
) UseCases {
	return UseCases{
		AppLicenses:  NewAppLicenses(sysUser, findAllAppLicences),
		UserLicenses: NewUserLicenses(sysUser, findAllUserLicences, countAssignedUserLicense),
	}
}
