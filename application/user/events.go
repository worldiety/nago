// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"go.wdy.de/nago/application/license"
	"golang.org/x/text/language"
	"time"
)

type Created struct {
	ID                ID
	Firstname         string
	Lastname          string
	Email             Email
	PreferredLanguage language.Tag
	NotifyUser        bool
	VerificationCode  Code
}

type LicensesUpdated struct {
	ID        ID
	Firstname string
	Lastname  string
	Email     Email
	Licenses  []license.ID
}

type MFACodeCreated struct {
	ID                ID
	Firstname         string
	Lastname          string
	Email             Email
	PreferredLanguage language.Tag
	ValidUntil        time.Time
	Code              string
}
