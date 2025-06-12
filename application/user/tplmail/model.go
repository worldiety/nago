// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tplmail

import (
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	"golang.org/x/text/language"
)

type MailVerificationModel struct {
	ID                user.ID
	Title             string
	Firstname         string
	Lastname          string
	Email             user.Email
	PreferredLanguage language.Tag
	ConfirmURL        core.URI
	ApplicationName   string
}

type PasswordResetModel struct {
	ID                user.ID
	Title             string
	Firstname         string
	Lastname          string
	Email             user.Email
	PreferredLanguage language.Tag
	ConfirmURL        core.URI
	ApplicationName   string
}
