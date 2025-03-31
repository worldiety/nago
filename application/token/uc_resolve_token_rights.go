// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package token

import (
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"os"
	"slices"
	"strings"
)

func NewResolveTokenRights(repo Repository, findGroupByID group.FindByID, findRoleByID role.FindByID, findUserByID user.FindByID, findLicenseByID license.FindUserLicenseByID) ResolveTokenRights {
	return func(subject auth.Subject, id ID) (ResolvedTokenRights, error) {
		if !subject.Valid() {
			return ResolvedTokenRights{}, user.InvalidSubjectErr
		}

		optToken, err := repo.FindByID(id)
		if err != nil {
			return ResolvedTokenRights{}, err
		}

		if optToken.IsNone() {
			return ResolvedTokenRights{}, os.ErrNotExist
		}

		token := optToken.Unwrap()
		allowedToView := subject.HasResourcePermission(repo.Name(), string(token.ID), PermResolveTokenRights) || token.Impersonation.UnwrapOr("") == subject.ID()
		if !allowedToView {
			return ResolvedTokenRights{}, user.PermissionDeniedErr
		}

		var res ResolvedTokenRights

		var groups []group.ID
		var roles []role.ID
		var permissions []permission.ID
		var licenses []license.ID

		// security note: always keep either or implementation for impersonation
		if uid := token.Impersonation.UnwrapOr(""); uid != "" {
			optUsr, err := findUserByID(user.SU(), uid)
			if err != nil {
				return ResolvedTokenRights{}, err
			}

			if optUsr.IsNone() {
				// user is gone
				return ResolvedTokenRights{}, nil
			}

			usr := optUsr.Unwrap()

			groups = usr.Groups
			roles = usr.Roles
			permissions = usr.Permissions
			licenses = usr.Licenses

			res.Impersonated = true
		} else {
			groups = token.Groups
			roles = token.Roles
			permissions = token.Permissions
			licenses = token.Licenses
		}

		// now start actual resolving the grants

		for _, gid := range groups {
			optGrp, err := findGroupByID(user.SU(), gid)
			if err != nil {
				return ResolvedTokenRights{}, err
			}

			if optGrp.IsNone() {
				continue
			}

			res.Groups = append(res.Groups, optGrp.Unwrap())
		}

		for _, rid := range roles {
			optRole, err := findRoleByID(user.SU(), rid)
			if err != nil {
				return ResolvedTokenRights{}, err
			}

			if optRole.IsNone() {
				continue
			}

			res.Roles = append(res.Roles, optRole.Unwrap())
		}

		for _, lic := range licenses {
			optLicense, err := findLicenseByID(user.SU(), lic)
			if err != nil {
				return ResolvedTokenRights{}, err
			}

			if optLicense.IsNone() {
				continue
			}

			res.Licenses = append(res.Licenses, optLicense.Unwrap())
		}

		for _, pid := range permissions {
			perm, ok := permission.Find(pid)
			if !ok {
				continue
			}

			res.Permissions = append(res.Permissions, perm)
		}

		for _, r := range res.Roles {
			for _, p := range r.Permissions {
				perm, ok := permission.Find(p)
				if !ok {
					continue
				}

				res.Permissions = append(res.Permissions, perm)
			}
		}

		// custom sorting and compacting
		slices.SortFunc(res.Roles, func(a, b role.Role) int {
			return strings.Compare(a.Name, b.Name)
		})

		slices.SortFunc(res.Groups, func(a, b group.Group) int {
			return strings.Compare(a.Name, b.Name)
		})

		slices.SortFunc(res.Roles, func(a, b role.Role) int {
			return strings.Compare(a.Name, b.Name)
		})

		slices.SortFunc(res.Permissions, func(a, b permission.Permission) int {
			return strings.Compare(a.Name, b.Name)
		})

		res.Permissions = slices.Compact(res.Permissions)

		return res, nil
	}
}
