// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/image"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/pkg/data"
	"golang.org/x/crypto/sha3"
)

func NewMergeSingleSignOnUser(mutex *sync.Mutex, repo Repository, findByMail FindByMail, loadGlobal settings.LoadGlobal, createSrcSet image.CreateSrcSet, rdb *rebac.DB) MergeSingleSignOnUser {
	return func(createData SingleSignOnUser, avatarBuf []byte) (ID, error) {
		mutex.Lock()
		defer mutex.Unlock()

		cfg := settings.ReadGlobal[Settings](loadGlobal)

		createData.Email = Email(strings.ToLower(string(createData.Email)))
		if !createData.Email.Valid() {
			return "", fmt.Errorf("email is invalid: %s", createData.Email)
		}

		optUser, err := findByMail(SU(), createData.Email)
		if err != nil {
			return "", fmt.Errorf("cannot find user by mail: %w", err)
		}

		if optUser.IsNone() {
			id := data.RandIdent[ID]()
			if optUsr, err := repo.FindByID(id); err != nil || optUsr.IsSome() {
				if err != nil {
					return "", fmt.Errorf("cannot find user by id: %w", err)
				}

				if optUser.IsSome() {
					return "", fmt.Errorf("random id collision: %s: %w", id, os.ErrExist)
				}
			}

			usr := User{
				ID:             id,
				NLSManagedUser: true,
				Email:          createData.Email,
				Contact: Contact{
					Firstname:         createData.FirstName(),
					Lastname:          createData.LastName(),
					MobilePhone:       createData.MobilePhone,
					Country:           createData.Country,
					State:             createData.State,
					PostalCode:        createData.PostalCode,
					City:              createData.City,
					Position:          createData.Position,
					ProfessionalGroup: createData.ProfessionalGroup,
					CompanyName:       createData.CompanyName,
					DisplayLanguage:   createData.PreferredLanguage,
					AboutMe:           createData.AboutMe,
				},
				CreatedAt:     time.Now(),
				EMailVerified: true,
				Status:        Enabled{},
			}

			if err := repo.Save(usr); err != nil {
				return "", fmt.Errorf("cannot save user: %w", err)
			}

			for _, rid := range cfg.DefaultRoles {
				err := rdb.Put(rebac.Triple{
					Source: rebac.Entity{
						Namespace: Namespace,
						Instance:  rebac.Instance(usr.ID),
					},
					Relation: rebac.Member,
					Target: rebac.Entity{
						Namespace: role.Namespace,
						Instance:  rebac.Instance(rid),
					},
				})

				if err != nil {
					return "", err
				}
			}

			for _, gid := range cfg.DefaultGroups {
				err := rdb.Put(rebac.Triple{
					Source: rebac.Entity{
						Namespace: Namespace,
						Instance:  rebac.Instance(usr.ID),
					},
					Relation: rebac.Member,
					Target: rebac.Entity{
						Namespace: group.Namespace,
						Instance:  rebac.Instance(gid),
					},
				})

				if err != nil {
					return "", err
				}
			}

			// done
			return usr.ID, nil
		}

		// merge existing
		user := optUser.Unwrap()

		if len(avatarBuf) > 0 {
			h := sha3.Sum256(avatarBuf)
			id := image.ID(hex.EncodeToString(h[:]))
			srcSet, err := createSrcSet(SU(), image.Options{
				ID: id,
			}, image.MemFile{
				Filename: "avatar.png",
				Bytes:    avatarBuf,
			})

			if err != nil {
				slog.Error("failed to generate nls user avatar image", "user", user.ID, "err", err.Error())
			} else {
				user.Contact.Avatar = srcSet.ID
			}
		}

		// TODO this is incomplete and may be we need some advanced merge logic?
		user.Contact.Firstname = createData.FirstName()
		user.Contact.Lastname = createData.LastName()
		user.Contact.MobilePhone = createData.MobilePhone
		user.Contact.Country = createData.Country
		user.Contact.State = createData.State
		user.Contact.PostalCode = createData.PostalCode
		user.Contact.City = createData.City
		user.Contact.Position = createData.Position
		user.Contact.ProfessionalGroup = createData.ProfessionalGroup
		user.Contact.CompanyName = createData.CompanyName
		user.Contact.DisplayLanguage = createData.PreferredLanguage
		user.Contact.AboutMe = createData.AboutMe

		// clear auth status
		user.NLSManagedUser = true
		user.VerificationCode = Code{}
		user.EMailVerified = true
		user.PasswordRequestCode = Code{}
		user.PasswordHash = nil
		user.RequirePasswordChange = false

		// security note: DO NEVER merge default roles or groups. People keep requesting for that,
		// but we must keep rejecting such a feature request because it is highly dangerous:
		// If an existing SSO user has been removed from groups and roles explicitly, we would
		// add them back to the default groups and roles, which would become a serious security incident.

		if err := repo.Save(user); err != nil {
			return "", fmt.Errorf("cannot save user: %w", err)
		}

		return user.ID, nil
	}
}
