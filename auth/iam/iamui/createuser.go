package iamui

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/uix/xform"
)

type createUser struct {
	Firstname string
	Lastname  string
	EMail     string
	Password1 string
	Password2 string
}

func create(subject auth.Subject, modals ui.ModalOwner, users *iam.Service) {
	var model createUser
	b := xform.NewBinding()
	xform.String(b, &model.Firstname, xform.Field{Label: "Vorname"})
	xform.String(b, &model.Lastname, xform.Field{Label: "Nachname"})
	mail := xform.String(b, &model.EMail, xform.Field{Label: "eMail"})
	pwd1 := xform.PasswordString(b, &model.Password1, xform.Field{Label: "Kennwort"})
	pwd2 := xform.PasswordString(b, &model.Password2, xform.Field{Label: "Kennwort wiederholen"})

	xform.Show(modals, b, func() error {
		if !iam.Email(mail.Value().Get()).Valid() {
			mail.Error().Set("Die eMail-Adresse ist ungültig.")
			return xform.UserMustCorrectInput
		}

		if model.Password1 != model.Password2 {
			msg := "Die Kennwörter stimmen nicht überein."
			pwd1.Error().Set(msg)
			pwd2.Error().Set(msg)
			return xform.UserMustCorrectInput
		}

		_, err := users.NewUser(subject, model.EMail, model.Firstname, model.Lastname, model.Password1)
		if err != nil {
			b.SetError(err.Error())
			return xform.UserMustCorrectInput
		}

		return nil
	})
}
