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
	"sync"
	"time"

	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/image"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/std"
	"golang.org/x/crypto/sha3"
)

// findSingleSignOnUser resolves the local user for the given external identity.
//
// The stable subject id of the identity provider is the primary criteria, because in contrast to the mail
// address it never changes. Only if it is unknown, we fall back to the mail address, which also backfills the
// subject id for accounts that have been merged before this field existed.
func findSingleSignOnUser(repo Repository, idx *UserIndex, createData SingleSignOnUser) (std.Option[User], error) {
	if id, ok, err := idx.LookupNLSUserID(createData.ID); err != nil {
		return std.None[User](), fmt.Errorf("cannot lookup nls user id: %w", err)
	} else if ok {
		optUsr, err := repo.FindByID(id)
		if err != nil {
			return std.None[User](), fmt.Errorf("cannot find user by id: %w", err)
		}

		if optUsr.IsSome() {
			return optUsr, nil
		}

		// stale index entry, fall through to the mail based matching
		slog.Warn("nls user id index points to a missing user", "nlsUserId", createData.ID, "user", id)
	}

	id, ok, err := idx.LookupMail(createData.Email)
	if err != nil {
		return std.None[User](), fmt.Errorf("cannot lookup mail: %w", err)
	}

	if !ok {
		return std.None[User](), nil
	}

	optUsr, err := repo.FindByID(id)
	if err != nil {
		return std.None[User](), fmt.Errorf("cannot find user by id: %w", err)
	}

	if optUsr.IsNone() {
		return std.None[User](), nil
	}

	usr := optUsr.Unwrap()

	// security note: the mail address alone must never be enough to take over an account which already
	// belongs to a different external identity, otherwise anybody who manages to get that address assigned
	// within any connected identity provider could hijack it. This can legitimately happen after a tenant
	// migration, which an administrator has to resolve by hand.
	if createData.ID != "" && usr.NLSUserID != "" && usr.NLSUserID != createData.ID {
		slog.Error("refused nls login, mail belongs to a different external identity",
			"user", usr.ID, "mail", createData.Email, "known", usr.NLSUserID, "provided", createData.ID)

		return std.None[User](), fmt.Errorf("mail %s belongs to the external identity %s, but %s was provided: %w",
			createData.Email, usr.NLSUserID, createData.ID, os.ErrPermission)
	}

	return optUsr, nil
}

func NewMergeSingleSignOnUser(mutex *sync.Mutex, bus events.Bus, repo Repository, idx *UserIndex, loadGlobal settings.LoadGlobal, createSrcSet image.CreateSrcSet, rdb *rebac.DB) MergeSingleSignOnUser {
	return func(createData SingleSignOnUser, avatarBuf []byte) (ID, error) {
		mutex.Lock()
		defer mutex.Unlock()

		cfg := settings.ReadGlobal[Settings](loadGlobal)

		createData.Email = NormalizeEmail(createData.Email)
		if !createData.Email.Valid() {
			return "", fmt.Errorf("email is invalid: %s", createData.Email)
		}

		optUser, err := findSingleSignOnUser(repo, idx, createData)
		if err != nil {
			return "", err
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
				NLSUserID:      createData.ID,
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
						Namespace: role.Namespace,
						Instance:  rebac.Instance(rid),
					},
					Relation: rebac.Member,
					Target: rebac.Entity{
						Namespace: Namespace,
						Instance:  rebac.Instance(usr.ID),
					},
				})

				if err != nil {
					return "", err
				}
			}

			for _, gid := range cfg.DefaultGroups {
				err := rdb.Put(rebac.Triple{
					Source: rebac.Entity{
						Namespace: group.Namespace,
						Instance:  rebac.Instance(gid),
					},
					Relation: rebac.Member,
					Target: rebac.Entity{
						Namespace: Namespace,
						Instance:  rebac.Instance(usr.ID),
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

		// the mail address may have been changed within the identity provider. Because we matched by the
		// stable subject id, we can simply follow that change. The address stays verified, because the
		// identity provider has verified it, otherwise the user would be locked out immediately.
		if !user.Email.Equals(createData.Email) {
			oldMail := user.Email
			user, err = applyEmailChange(bus, repo, idx, user, createData.Email, emailChangeOptions{KeepVerified: true})
			if err != nil {
				return "", fmt.Errorf("cannot follow mail change of nls user %s from %s to %s: %w", user.ID, oldMail, createData.Email, err)
			}

			slog.Info("followed nls mail change", "user", user.ID, "old", oldMail, "new", createData.Email)
		}

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
		if createData.ID != "" {
			// note: never clear an already known subject id, an older NLS may simply not report it
			user.NLSUserID = createData.ID
		}
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
