package iamui

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/icon"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/uix/crud"
	"go.wdy.de/nago/presentation/uix/xform"
)

func Users(subject auth.Subject, owner ui.ModalOwner, service *iam.Service) core.Component {
	return crud.NewView(owner, crud.NewOptions[iam.User](func(opts *crud.Options[iam.User]) {
		opts.Title = "Nutzerkonten"
		opts.OnDelete(func(user iam.User) error {
			return service.DeleteUser(subject, user.ID)
		})
		opts.FindAll = service.AllUsers(subject)

		opts.OnUpdate(func(user iam.User) error {
			return service.UpdateUser(subject, user.ID, string(user.Email), user.Firstname, user.Lastname, user.Permissions, user.Roles, user.Groups)
		})

		if subject.HasPermission(iam.CreateUser) {
			opts.Actions = append(opts.Actions, ui.NewButton(func(btn *ui.Button) {
				btn.Caption().Set("Neuen Nutzer anlegen")
				btn.PreIcon().Set(icon.UserPlus)
				btn.Action().Set(func() {
					create(subject, owner, service)
				})
			}))
		}

		opts.Binding = crud.NewBinding[iam.User](func(bnd *crud.Binding[iam.User]) {
			crud.Text(bnd, crud.FromPtr("Vorname", func(model *iam.User) *string {
				return &model.Firstname
			}))
			crud.Text(bnd, crud.FromPtr("Nachname", func(model *iam.User) *string {
				return &model.Lastname
			}))
			crud.OneToMany(bnd, service.AllPermissions(subject), func(permission iam.Permission) string {
				return permission.Name()
			}, crud.FromPtr("Berechtigungen", func(model *iam.User) *[]iam.PID {
				return &model.Permissions
			}))

			crud.OneToMany(bnd,
				service.AllGroups(subject),
				func(usr iam.Group) string {
					return usr.Name
				},
				crud.FromPtr("Gruppen", func(model *iam.User) *[]auth.GID {
					return &model.Groups
				}),
			)

			crud.OneToMany(bnd,
				service.AllRoles(subject),
				func(usr iam.Role) string {
					return usr.Name
				},
				crud.FromPtr("Rollen", func(model *iam.User) *[]auth.RID {
					return &model.Roles
				}),
			)
		})
	}))
}

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
