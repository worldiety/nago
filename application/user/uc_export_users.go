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
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/consent"
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

	switch opts.Language {
	case language.German:
		writer.Comma = ';' // excel in germany does not use , by default
	}

	option.MustZero(writer.Write([]string{
		"id",
		"first_name",
		"last_name",
		"email",
		"email_verified",
		"created_at",
		"enabled",
		"roles",
		"consents",
		"salutation",
		"title",
		"position",
		"professional_group",
		"company_name",
		"mobile_phone",
		"phone",
		"day_of_birth",
	}))

	for usr, err := range repo.FindAllByID(slices.Values(users)) {
		if err != nil {
			return nil, err
		}

		var consents []string
		for _, c := range usr.Consents {
			consents = append(consents, string(c.ID)+":"+strconv.FormatBool(c.Status() == consent.Approved))
		}

		option.MustZero(writer.Write([]string{
			string(usr.ID),
			usr.Contact.Firstname,
			usr.Contact.Lastname,
			string(usr.Email),
			strconv.FormatBool(usr.EMailVerified),
			usr.CreatedAt.Format(time.RFC3339),
			strconv.FormatBool(usr.Enabled()),
			string(xstrings.Join(usr.Roles, ",")),
			strings.Join(consents, ","),
			usr.Contact.Salutation,
			usr.Contact.Title,
			usr.Contact.Position,
			usr.Contact.ProfessionalGroup,
			usr.Contact.CompanyName,
			usr.Contact.MobilePhone,
			usr.Contact.Phone,
			usr.Contact.DayOfBirth.Time(time.Local).Format(xtime.GermanDate),
		}))
	}

	writer.Flush()
	return f.Bytes(), nil
}
