package tplmail

import (
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	"golang.org/x/text/language"
)

type MailRegisteredSubjectModel struct {
	ID                user.ID
	Firstname         string
	Lastname          string
	Email             user.Email
	PreferredLanguage language.Tag
	ConfirmURL        core.URI
	ApplicationName   string
}
