// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"fmt"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/pkg/data"
	"time"
)

func NewEnableBootstrapAdmin(repo Repository, system SysUser, userByMail FindByMail) EnableBootstrapAdmin {
	return func(aliveUntil time.Time, password Password) (ID, error) {
		optUsr, err := userByMail(system(), "admin@localhost")
		if err != nil {
			return "", err
		}

		if err := password.Validate(); err != nil {
			return "", err
		}

		var usr User

		if optUsr.IsSome() {
			usr = optUsr.Unwrap()
		} else {
			usr.ID = data.RandIdent[ID]() // a random admin user id makes some attacks impossible
			usr.Email = "admin@localhost"
			usr.Contact.Firstname = "admin"
			usr.Contact.Lastname = "admin"
		}

		usr.Status = EnabledUntil{ValidUntil: aliveUntil}

		// we are not allowed to have domain specific permissions, only those to bootstrap other users.
		// even admins must not see customers secret domain stuff.
		usr.Permissions = []permission.ID{
			PermCreate,
			PermFindByID,
			PermFindByMail,
			PermFindAll,
			PermChangeOtherPassword,
			PermDelete,
			PermUpdateOtherContact,
			PermUpdateOtherRoles,
			PermUpdateOtherPermissions,
			PermUpdateOtherLicenses,
			PermUpdateOtherGroups,
			PermCountAssignedUserLicense,
			PermRevokeAssignedUserLicense,
			PermUpdateaccountStatus,
			group.PermCreate,
			group.PermFindAll,
			group.PermDelete,
			group.PermFindByID,
			group.PermUpdate,
			role.PermCreate,
			role.PermFindByID,
			role.PermFindAll,
			role.PermDelete,
			role.PermUpdate,
			permission.PermFindAll,
			license.PermFindAllAppLicenses,
			license.PermFindAppLicenseByID,
			license.PermCreateAppLicense,
			license.PermUpdateAppLicense,
			license.PermDeleteAppLicense,
			license.PermFindAllUserLicenses,
			license.PermFindUserLicenseByID,
			license.PermCreateUserLicense,
			license.PermUpdateUserLicense,
			license.PermDeleteUserLicense,

			// avoid import cycles, thus hard code our bootstrap permissions
			"nago.backup.backup",
			"nago.backup.restore",
			"nago.billing.license.app",
			"nago.billing.license.user",

			"nago.secret.find_my_secrets",
			"nago.secret.create",
			"nago.secret.groups.update",
			"nago.secret.credentials.update",
			"nago.secret.delete",

			"nago.mail.send",
			"nago.mail.outgoing.find_by_id",
			"nago.mail.outgoing.find_all",
			"nago.mail.outgoing.create",
			"nago.mail.outgoing.update",
			"nago.mail.outgoing.delete_by_id",

			"nago.template.execute",
			"nago.template.find_all",
			"nago.template.create",
			"nago.template.ensure_build_in",

			"nago.settings.global.load",
			"nago.settings.global.store",

			"nago.template.execute",
			"nago.template.find_all",
			"nago.template.find_by_id",
			"nago.template.project.blob.load",
			"nago.template.project.blob.update",
			"nago.template.project.blob.delete",
			"nago.template.project.blob.rename",
			"nago.template.project.blob.create",
			"nago.template.project.runcfg.add",
			"nago.template.project.runcfg.remove",
			"nago.template.project.export",
			"nago.template.project.import",
			"nago.template.create",
			"nago.template.delete",
			"nago.template.ensure_build_in",

			// security note: we do not allow the masterkey permission by default, they are to dangerous
			//nago.backup.masterkey.export
			//nago.backup.masterkey.replace

			"nago.data.inspector",
		}

		hType := Argon2IdMin
		salt, hash, err := password.Hash(hType)
		if err != nil {
			return "", fmt.Errorf("hashing password: %w", err)
		}

		usr.LastPasswordChangedAt = time.Now()
		usr.EMailVerified = true
		usr.Salt = salt
		usr.PasswordHash = hash
		usr.Algorithm = hType

		if err := repo.Save(usr); err != nil {
			return "", fmt.Errorf("cannot save bootstrap user: %w", err)
		}

		return usr.ID, nil
	}
}
