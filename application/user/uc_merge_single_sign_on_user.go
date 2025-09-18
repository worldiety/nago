// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"go.wdy.de/nago/pkg/data"
)

func NewMergeSingleSignOnUser(mutex *sync.Mutex, repo Repository, findByMail FindByMail) MergeSingleSignOnUser {
	return func(createData SingleSignOnUser) (ID, error) {
		mutex.Lock()
		defer mutex.Unlock()

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

			// done
			return usr.ID, nil
		}

		// merge existing
		user := optUser.Unwrap()

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

		if err := repo.Save(user); err != nil {
			return "", fmt.Errorf("cannot save user: %w", err)
		}

		return user.ID, nil
	}
}
