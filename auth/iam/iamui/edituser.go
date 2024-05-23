package iamui

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/uix/xdialog"
	"go.wdy.de/nago/presentation/uix/xform"
)

type editUserModel struct {
	Firstname   string
	Lastname    string
	EMail       string
	Permissions []iam.PID
}

func editUser(subject auth.Subject, modals ui.ModalOwner, id auth.UID, users *iam.Service) {

	optUsr, err := users.FindUser(subject, id)
	if xdialog.HandleError(modals, "interner Fehler", err) {
		return
	}

	if !optUsr.Valid {
		xdialog.ShowMessage(modals, "Der Nutzer existiert nicht.")
		return
	}

	usr := optUsr.Unwrap()

	var model editUserModel
	model.Firstname = usr.Firstname
	model.Lastname = usr.Lastname
	model.EMail = string(usr.Email)
	model.Permissions = usr.Permissions

	b := xform.NewBinding()
	xform.String(b, &model.Firstname, xform.Field{Label: "Vorname"})
	xform.String(b, &model.Lastname, xform.Field{Label: "Nachname"})
	xform.String(b, &model.EMail, xform.Field{Label: "eMail"})
	xform.OneToMany(b, &model.Permissions, users.AllPermissions(subject), func(e iam.Permission) string {
		return e.Name()
	}, xform.Field{Label: "Berechtigungen", Hint: "Konfiguration der individuellen Einzelberechtigungen. Weitere Berechtigungen aus Gruppen und Rollen werden vererbt, die hier nicht zu sehen sind."})

	xform.Show(modals, b, func() error {

		err := users.UpdateUser(subject, id, model.EMail, model.Firstname, model.Lastname, model.Permissions)
		if err != nil {
			b.SetError(err.Error())
			return xform.UserMustCorrectInput
		}

		return nil
	})
}
