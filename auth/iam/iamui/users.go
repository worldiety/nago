package iamui

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/icon"
	"go.wdy.de/nago/presentation/uilegacy"
	"go.wdy.de/nago/presentation/uix/crud"
	"go.wdy.de/nago/presentation/uix/xform"
	"log/slog"
	"strings"
)

func Users(wnd core.Window, owner uilegacy.ModalOwner, service *iam.Service) core.View {
	subject := wnd.Subject()
	return crud.NewView(owner, crud.NewOptions[iam.User](func(opts *crud.Options[iam.User]) {
		opts.Title("Nutzerkonten")
		opts.Delete(func(user iam.User) error {
			return service.DeleteUser(subject, user.ID)
		})
		opts.ReadAll(service.AllUsers(subject))

		opts.Update(func(user iam.User) error {
			return service.UpdateUser(subject, user.ID, string(user.Email), user.Firstname, user.Lastname, user.Permissions, user.Roles, user.Groups)
		})

		opts.Responsive(wnd)

		if subject.HasPermission(iam.CreateUser) {
			opts.Actions(uilegacy.NewButton(func(btn *uilegacy.Button) {
				btn.Caption().Set("Neuen Nutzer anlegen")
				btn.PreIcon().Set(icon.UserPlus)
				btn.Action().Set(func() {
					create(subject, owner, service)
				})
			}))
		}

		opts.Bind(func(bnd *crud.Binding[iam.User]) {
			crud.Text(bnd, crud.FromPtr("Vorname", func(model *iam.User) *string {
				return &model.Firstname
			}))
			crud.Text(bnd, crud.FromPtr("Nachname", func(model *iam.User) *string {
				return &model.Lastname
			}))
			crud.OneToMany(bnd, service.AllPermissions(subject), func(permission iam.Permission) string {
				return permission.Name()
			}, crud.Field[iam.User, []iam.PID]{
				Caption: "Berechtigungen",
				Stringer: func(user iam.User) string {
					// this is by intention: we show the entire inherited list of all permission, not just the customized ones
					perms, err := service.FindAllUserPermissions(subject, user.ID)
					if err != nil {
						slog.Error("failed to find all user permissions", "err", err, "uid", user.ID)
					}

					actualPerms := service.AllPermissionsByIDs(subject, perms...)
					var tmp []string
					actualPerms(func(permission iam.Permission, e error) bool {
						if e != nil {
							err = e
							return false
						}

						tmp = append(tmp, permission.Name())
						return true
					})

					return strings.Join(tmp, ", ")
				},
				FromModel: func(user iam.User) []iam.PID {
					return user.Permissions // this is by intention, only allow editing the custom ones
				},
				IntoModel: func(model iam.User, value []iam.PID) (iam.User, error) {
					model.Permissions = value // this is by intention, only allow editing the custom ones
					return model, nil
				},
			})

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

func create(subject auth.Subject, modals uilegacy.ModalOwner, users *iam.Service) {
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
