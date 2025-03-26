// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"fmt"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/std"
	"golang.org/x/text/language"
	"log/slog"
	"strings"
	"sync"
	"time"
)

func NewCreate(mutex *sync.Mutex, loadGlobal settings.LoadGlobal, eventBus events.EventBus, findByMail FindByMail, repo Repository) Create {
	return func(subject permission.Auditable, model ShortRegistrationUser) (User, error) {
		if err := subject.Audit(PermCreate); err != nil {
			return User{}, err
		}

		// this is really harsh and allows intentionally only to create one user per second
		mutex.Lock()
		defer mutex.Unlock()

		requiredPasswordChange := false
		if model.Password == "" && model.PasswordRepeated == "" {
			model.Password = data.RandIdent[Password]()
			model.PasswordRepeated = model.Password
			requiredPasswordChange = true
		}

		if model.Password != model.PasswordRepeated {
			return User{}, std.NewLocalizedError("Eingabebeschränkung", "Die Kennwörter stimmen nicht überein.")
		}

		mail := Email(strings.TrimSpace(strings.ToLower(string(model.Email))))
		if !mail.Valid() {
			return User{}, std.NewLocalizedError("Eingabebeschränkung", "Auch wenn es sich um eine potentiell gültige E-Mail Adresse handeln könnte, wird dieses Format nicht unterstützt.")
		}

		if err := model.Password.Validate(); err != nil {
			return User{}, err
		}

		salt, hash, err := model.Password.Hash(Argon2IdMin)
		if err != nil {
			return User{}, err
		}

		createdAt := time.Now()
		user := User{
			// see https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html#user-ids
			ID:        data.RandIdent[ID](),
			Email:     mail,
			Algorithm: Argon2IdMin,
			Contact: Contact{
				Title:             model.Title,
				Firstname:         model.Firstname,
				Lastname:          model.Lastname,
				Country:           model.Country,
				City:              model.City,
				PostalCode:        model.PostalCode,
				State:             model.State,
				Position:          model.Position,
				ProfessionalGroup: model.ProfessionalGroup,
				CompanyName:       model.CompanyName,
				DisplayLanguage:   model.PreferredLanguage.String(),
			},
			Salt:                  salt,
			PasswordHash:          hash,
			CreatedAt:             createdAt,
			LastPasswordChangedAt: createdAt,
			Status:                Enabled{},
			EMailVerified:         model.Verified,
			RequirePasswordChange: requiredPasswordChange,
			// initially, give the user a week to respond. Note, that for self registration we just may
			// remove users which have never been verified automatically
			VerificationCode: NewCode(DefaultVerificationLifeTime),

			// legal fields
			Newsletter:                model.Newsletter,
			GeneralTermsAndConditions: model.GeneralTermsAndConditions,
			DataProtectionProvision:   model.DataProtectionProvision,
			MinAge:                    model.MinAge,
		}

		if model.SelfRegistered {
			userSettings := settings.ReadGlobal[Settings](loadGlobal)
			if !userSettings.SelfRegistration {
				return User{}, fmt.Errorf("self registration is not allowed")
			}

			if len(userSettings.AllowedDomains) > 0 {
				allowed := false
				for _, allowedPostfix := range userSettings.AllowedDomains {
					if strings.HasSuffix(strings.ToLower(string(user.Email)), strings.ToLower(strings.TrimSpace(allowedPostfix))) {
						allowed = true
						break
					}
				}

				if !allowed {
					return User{}, std.NewLocalizedError("Registrierung nicht möglich", "Sie dürfen sich leider nicht registrieren.")
				}
			}

			user.Roles = append(user.Roles, userSettings.DefaultRoles...)
			user.Groups = append(user.Groups, userSettings.DefaultGroups...)
		}

		// intentionally validate now, so that an attacker cannot use this method to massively
		// find out, which mails exist in the system
		optView, err := findByMail(subject, mail)
		if err != nil {
			return User{}, fmt.Errorf("cannot check for existing user: %w", err)
		}

		if optView.IsSome() {
			// security note: this allows to expose the fact, that a user already exists in the system.
			// however, there is no reasonable way to avoid that.
			// We must not delay this with a sleep, because we are still in the write mutex lock and therefore
			// would cause a kind of accumulating "deadlock"/DOS like behavior, for any other use cases
			// which need the lock.
			return User{}, std.NewLocalizedError("Nutzerregistrierung", "Die E-Mail-Adresse wird bereits verwendet.")
		}

		// unlikely, but better safe than sorry
		optUsr, err := repo.FindByID(user.ID)
		if err != nil {
			return User{}, fmt.Errorf("cannot find user by id: %w", err)
		}

		if optUsr.IsSome() {
			return User{}, fmt.Errorf("user id already taken")
		}

		// persist
		err = repo.Save(user)
		if err != nil {
			return User{}, fmt.Errorf("cannot persist new user: %w", err)
		}

		tag, err := language.Parse(user.Contact.DisplayLanguage)
		if err != nil {
			slog.Error("user contact has invalid preferred language", "err", err)
		}

		// publish in any case
		eventBus.Publish(Created{
			ID:                user.ID,
			Firstname:         user.Contact.Firstname,
			Lastname:          user.Contact.Lastname,
			Email:             user.Email,
			PreferredLanguage: tag,
			NotifyUser:        model.NotifyUser,
			VerificationCode:  user.VerificationCode,
		})

		return user, nil
	}
}
