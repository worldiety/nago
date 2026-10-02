// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tplmail

import (
	"embed"
	"go.wdy.de/nago/application/template"
)

const ID template.ID = "nago.template.system.mails"

const (
	MailVerification        template.DefinedTemplateName = "MailVerification"
	MailVerificationSubject template.DefinedTemplateName = "MailVerificationSubject"

	ResetPassword        template.DefinedTemplateName = "ResetPassword"
	ResetPasswordSubject template.DefinedTemplateName = "ResetPasswordSubject"

	OnboardingCode        template.DefinedTemplateName = "OnboardingCode"
	OnboardingCodeSubject template.DefinedTemplateName = "OnboardingCodeSubject"
)

//go:embed *.gohtml
var Files embed.FS
