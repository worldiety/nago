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
)

//go:embed *
var Files embed.FS
