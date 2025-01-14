package tplmail

import (
	"embed"
	"go.wdy.de/nago/application/template"
)

const ID template.ID = "nago.template.system.mails"

const MailRegistered template.DefinedTemplateName = "MailRegistered"
const MailRegisteredSubject template.DefinedTemplateName = "MailRegisteredSubject"

//go:embed *
var Files embed.FS
