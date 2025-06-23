// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package userimporter

import (
	"context"
	"fmt"
	"github.com/worldiety/jsonptr"
	"go.wdy.de/nago/application/consent"
	"go.wdy.de/nago/application/dataimport/importer"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/xstrings"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"iter"
	"reflect"
)

const (
	ID importer.ID = "nago.data.importer.user"
)

type User struct {
	ID    user.ID    `json:"id,omitempty" supportingText:"ID bezeichnet hier das intern technische Identifizierungsmerkmal. Es ist empfohlen, dies leer zu lassen, um ein eindeutiges Kennzeichen automatisch zu vergeben."`
	Email user.Email `json:"email,omitempty"`

	// Title incorporates the official title like Professor, Oberbürgermeister etc. but also
	// an academic degree like Diploma, Bachelor, Master or Doctor
	Title string `json:"title,omitempty" label:"Titel"`
	// Saluation is like Mr, Mrs or divers
	Salutation  string `json:"salutation,omitempty"`
	Firstname   string `json:"firstname,omitempty" label:"Vorname"`
	Lastname    string `json:"lastname,omitempty" label:"Nachname"`
	Phone       string `json:"phone,omitempty"`
	MobilePhone string `json:"mobilePhone,omitempty"`
	// Country is like Deutschland, not the BCP47 code
	Country    string `json:"country,omitempty"`
	City       string `json:"city,omitempty"`
	PostalCode string `json:"postalCode,omitempty"`
	State      string `json:"state,omitempty"`
	LinkedIn   string `json:"linkedIn,omitempty"`
	Website    string `json:"website,omitempty"`
	// Position is like CEO
	Position          string `json:"position,omitempty"`
	ProfessionalGroup string `json:"professionalGroup,omitempty"`
	CompanyName       string `json:"company,omitempty"`
	// DisplayLanguage is a BCP47 string like de or en_US of what the User wants to see its content.
	DisplayLanguage string       `json:"displayLanguage,omitempty"`
	AboutMe         string       `json:"aboutMe,omitempty"`
	Consents        []consent.ID `json:"consents,omitempty" source:"nago.consent.options"`
}

type usrImporter struct {
	users user.UseCases
}

func NewImporter(users user.UseCases) importer.Importer {
	return usrImporter{users: users}
}

func (u usrImporter) Identity() importer.ID {
	return ID
}

func (u usrImporter) Import(ctx context.Context, opts importer.Options, data iter.Seq2[*jsonptr.Obj, error]) error {
	for obj, err := range data {
		if err != nil {
			return err
		}

		usr, err := importer.FromJSON[User](obj)
		if err != nil {
			return fmt.Errorf("cannot convert to user: %v", err)
		}

		if usr.ID != "" {
			optUsr, err := u.users.FindByID(user.SU(), usr.ID)
			if err != nil {
				return fmt.Errorf("cannot find user: %v", err)
			}

			if optUsr.IsSome() {
				// TODO implement me
				if opts.ContinueOnError {
					continue
				}
				return fmt.Errorf("merging an existing user by id is currently not supported")
			}
		}

		if usr.Email != "" {
			optUsr, err := u.users.FindByID(user.SU(), usr.ID)
			if err != nil {
				return fmt.Errorf("cannot find user: %v", err)
			}

			if optUsr.IsSome() {
				// TODO implement me
				if opts.ContinueOnError {
					continue
				}
				return fmt.Errorf("merging an existing user by email is currently not supported")
			}
		}

		if usr.Email == "" {
			if opts.ContinueOnError {
				continue
			}
			return fmt.Errorf("user email is required")
		}

		_, err = u.users.Create(user.SU(), user.ShortRegistrationUser{
			Firstname:         usr.Firstname,
			Lastname:          usr.Lastname,
			Email:             usr.Email,
			NotifyUser:        false,
			Verified:          false,
			Title:             usr.Title,
			Position:          usr.Position,
			CompanyName:       usr.CompanyName,
			City:              usr.City,
			PostalCode:        usr.PostalCode,
			State:             usr.State,
			Country:           usr.Country,
			ProfessionalGroup: usr.ProfessionalGroup,
			MobilePhone:       usr.MobilePhone,
		})

		if err != nil {
			if opts.ContinueOnError {
				continue
			}
			return fmt.Errorf("cannot create user: %w", err)
		}
	}

	return nil
}

func (u usrImporter) Validate(ctx context.Context, obj *jsonptr.Obj) error {
	usr, err := importer.FromJSON[User](obj)
	if err != nil {
		return err
	}

	if !usr.Email.Valid() {
		return std.NewLocalizedError("Ungültige EMail", fmt.Sprintf("Die Mail-Adresse '%s' ist ungültig bzw. wird vom System nicht unterstützt", usr.Email))
	}

	return nil
}

func (u usrImporter) FindMatches(ctx context.Context, opts importer.MatchOptions, obj *jsonptr.Obj) iter.Seq2[importer.Match, error] {
	return func(yield func(importer.Match, error) bool) {
		for usr, err := range u.users.FindAll(user.SU()) {
			if err != nil {
				if !yield(importer.Match{}, err) {
					return
				}

				continue
			}

			if id, ok := obj.Get("id"); ok {
				if id.String() == string(usr.ID) {
					if !yield(importer.ToMatch(usr, 1)) {
						return
					}

					continue
				}
			}

			if mail, ok := obj.Get("email"); ok {
				if mail.String() == string(usr.Email) {
					if !yield(importer.ToMatch(usr, 0.9)) {
						return
					}

					continue
				}
			}

			usrName := xstrings.Join2(" ", usr.Contact.Firstname, usr.Contact.Lastname)
			a, _ := obj.Get("firstname")
			b, _ := obj.Get("lastname")
			otherName := xstrings.Join2(" ", a.String(), b.String())
			if score := importer.Similarity(usrName, otherName); score > 0.5 {
				if !yield(importer.ToMatch(usr, score)) {
					return
				}
			}

		}

	}
}

func (u usrImporter) Configuration() importer.Configuration {
	return importer.Configuration{
		Image:        icons.UserAdd,
		Name:         "Nutzer",
		Description:  "Nutzer aus externen Quellen importieren.",
		ExpectedType: reflect.TypeFor[User](),
		PreviewMappings: []importer.PreviewMapping{
			{Name: "Vorname", Keywords: []string{"/contact/firstname", "firstname"}},
			{Name: "Nachname", Keywords: []string{"/contact/lastname", "lastname", "name"}},
			{Name: "Firma", Keywords: []string{"/contact/company", "company", "unternehmen"}},
			{Name: "E-Mail", Keywords: []string{"/email", "email", "mail"}},
		},
	}
}
