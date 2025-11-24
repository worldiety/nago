// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/consent"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/pkg/xstrings"
	"go.wdy.de/nago/pkg/xtime"
	"golang.org/x/text/language"
)

func NewExportUsers(repo Repository) ExportUsers {
	return func(subject Subject, users []ID, opts ExportUsersOptions) ([]byte, error) {
		if err := subject.Audit(PermExportUsers); err != nil {
			return nil, err
		}

		switch opts.Format {
		case ExportCSV:
			return createCSV(repo, users, opts)
		default:
			return nil, fmt.Errorf("unsupported format: %v", opts.Format)
		}
	}
}

func createCSV(repo Repository, users []ID, opts ExportUsersOptions) ([]byte, error) {
	var f bytes.Buffer
	option.Must(f.Write([]byte{0xEF, 0xBB, 0xBF})) // unicode BOM for excel

	writer := csv.NewWriter(&f)

	switch opts.Language.Parent() {
	case language.German:
		writer.Comma = ';' // excel in germany does not use , by default
	}

	columns := []string{
		"id",
		"first_name",
		"last_name",
		"email",
		"email_verified",
		"created_at",
		"enabled",
		"roles",
		"salutation",
		"title",
		"position",
		"professional_group",
		"company_name",
		"mobile_phone",
		"phone",
		"day_of_birth",
	}

	loadedUsers, err := xslices.Collect2(repo.FindAllByID(slices.Values(users)))
	if err != nil {
		return nil, err
	}

	// build all available consents
	allConsents := map[consent.ID]bool{}
	for _, usr := range loadedUsers {
		for _, c := range usr.Consents {
			allConsents[c.ID] = true
		}
	}

	sortedConsentIdents := slices.Sorted(maps.Keys(allConsents))
	for _, id := range sortedConsentIdents {
		columns = append(columns, string(id))
	}

	option.MustZero(writer.Write(columns))

	for _, usr := range loadedUsers {
		row := []string{
			string(usr.ID),
			usr.Contact.Firstname,
			usr.Contact.Lastname,
			string(usr.Email),
			strconv.FormatBool(usr.EMailVerified),
			usr.CreatedAt.Format(time.RFC3339),
			strconv.FormatBool(usr.Enabled()),
			string(xstrings.Join(usr.Roles, ",")),
			usr.Contact.Salutation,
			usr.Contact.Title,
			usr.Contact.Position,
			usr.Contact.ProfessionalGroup,
			usr.Contact.CompanyName,
			usr.Contact.MobilePhone,
			usr.Contact.Phone,
			usr.Contact.DayOfBirth.Time(time.Local).Format(xtime.GermanDate),
		}

		for _, ident := range sortedConsentIdents {
			row = append(row, strconv.FormatBool(usr.ConsentStatusByID(ident) == consent.Approved))
		}

		option.MustZero(writer.Write(row))
	}

	writer.Flush()
	return f.Bytes(), nil
}
