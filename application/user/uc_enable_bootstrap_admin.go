// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"fmt"
	"strings"
	"time"

	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/pkg/data"
)

func NewEnableBootstrapAdmin(repo Repository, system SysUser, userByMail FindByMail, rdb *rebac.DB) EnableBootstrapAdmin {
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
		// let us insert only nago.* permissions
		for perm := range permission.All() {
			if !strings.HasPrefix(string(perm.ID), "nago.") {
				continue
			}

			err := rdb.Put(rebac.Triple{
				Source: rebac.Entity{
					Namespace: Namespace,
					Instance:  rebac.Instance(usr.ID),
				},
				Relation: rebac.Relation(perm.ID),
				Target: rebac.Entity{
					Namespace: rebac.Global,
					Instance:  rebac.AllInstances,
				},
			})
			if err != nil {
				return "", fmt.Errorf("cannot add permission to bootstrap user: %w", err)
			}
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
