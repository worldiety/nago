package iamui

import (
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/uix/xdialog"
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
	xform.String(b, &model.EMail, xform.Field{Label: "eMail"})
	xform.PasswordString(b, &model.Password1, xform.Field{Label: "Kennwort"})
	xform.PasswordString(b, &model.Password2, xform.Field{Label: "Kennwort wiederholen"})

	xform.Show(modals, b, func() error {
		if model.Password1 != model.Password2 {
			xdialog.ShowMessage(modals, "Die Kennwörter stimmen nicht überein.")
			return fmt.Errorf("passwords don't match")
		}
		_, err := users.NewUser(subject, model.EMail, model.Firstname, model.Lastname, model.Password1)
		if err != nil {
			return err
		}

		return nil
	})
}
