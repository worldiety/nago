package iamui

import (
	"errors"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/crud"
	"log/slog"
)

func Users(wnd core.Window, service *iam.Service) core.View {
	subject := wnd.Subject()

	bnd := crud.NewBinding[iam.User](wnd)
	bnd.Add(
		crud.Text(crud.TextOptions{Label: "ID"}, crud.Ptr(func(e *iam.User) *auth.UID {
			return &e.ID
		})).ReadOnly(true).WithoutTable(),
		crud.Text(crud.TextOptions{Label: "Vorname"}, crud.Ptr(func(e *iam.User) *string {
			return &e.Firstname
		})),
		crud.Text(crud.TextOptions{Label: "Nachname"}, crud.Ptr(func(e *iam.User) *string {
			return &e.Lastname
		})),
		crud.OneToMany(crud.OneToManyOptions[iam.Permission, iam.PID]{
			Label:           "Vererbte Berechtigungen",
			ForeignEntities: service.AllPermissions(subject),
			ForeignPickerRenderer: func(t iam.Permission) core.View {
				return ui.Text(t.Name())
			},
		}, crud.Ptr(func(user *iam.User) *[]iam.PID {
			// this is by intention: we show the entire inherited list of all permission, not just the customized ones
			perms, err := service.FindAllUserPermissions(subject, user.ID)
			if err != nil {
				slog.Error("failed to find all user permissions", "err", err, "uid", user.ID)
			}

			return &perms
		})).ReadOnly(true),

		crud.OneToMany(crud.OneToManyOptions[iam.Permission, iam.PID]{
			Label:           "Einzelberechtigungen",
			ForeignEntities: service.AllPermissions(subject),
			ForeignPickerRenderer: func(t iam.Permission) core.View {
				return ui.Text(t.Name())
			},
		}, crud.Ptr(func(user *iam.User) *[]iam.PID {
			return &user.Permissions
		})),

		crud.OneToMany(crud.OneToManyOptions[iam.Group, auth.GID]{
			Label:           "Gruppen",
			ForeignEntities: service.AllGroups(subject),
			ForeignPickerRenderer: func(t iam.Group) core.View {
				return ui.Text(t.Name)
			},
		}, crud.Ptr(func(model *iam.User) *[]auth.GID {
			return &model.Groups
		})),

		crud.OneToMany(crud.OneToManyOptions[iam.Role, auth.RID]{
			Label:           "Rollen",
			ForeignEntities: service.AllRoles(subject),
			ForeignPickerRenderer: func(t iam.Role) core.View {
				return ui.Text(t.Name)
			},
		}, crud.Ptr(func(model *iam.User) *[]auth.RID {
			return &model.Roles
		})),

		crud.AggregateActions(
			"Optionen",
			crud.ButtonDelete(wnd, func(group iam.User) error {
				return service.DeleteUser(subject, group.ID)
			}),
			crud.ButtonEdit(bnd, func(user iam.User) (errorText string, infrastructureError error) {
				return "", service.UpdateUser(subject, user.ID, string(user.Email), user.Firstname, user.Lastname, user.Permissions, user.Roles, user.Groups)
			}),
		),
	)

	bndCrUsr := crud.NewBinding[createUser](wnd).Add(
		crud.Text(crud.TextOptions{Label: "Vorname"}, crud.Ptr(func(e *createUser) *string {
			return &e.Firstname
		})),
		crud.Text(crud.TextOptions{Label: "Nachname"}, crud.Ptr(func(e *createUser) *string {
			return &e.Lastname
		})),
		crud.Text(crud.TextOptions{Label: "eMail"}, crud.Ptr(func(e *createUser) *string {
			return &e.EMail
		})),
		crud.Password(crud.PasswordOptions{Label: "Kennwort"}, crud.Ptr(func(e *createUser) *string {
			return &e.Password1
		})),
		crud.Password(crud.PasswordOptions{Label: "Kennwort wiederholen"}, crud.Ptr(func(e *createUser) *string {
			return &e.Password2
		})),
	)
	opts := crud.Options(bnd).
		Actions(crud.ButtonCreate[createUser](bndCrUsr, createUser{}, func(model createUser) (errorText string, infrastructureError error) {
			if !iam.Email(model.EMail).Valid() {
				return "Die eMail-Adresse ist ungültig.", nil
			}

			if model.Password1 != model.Password2 {
				return "Die Kennwörter stimmen nicht überein.", nil
			}

			_, err := service.NewUser(subject, model.EMail, model.Firstname, model.Lastname, model.Password1)
			var pwdErr iam.WeakPasswordError
			if errors.As(err, &pwdErr) {
				return pwdErr.Error(), nil
			}

			return "", err
		})).
		ViewStyle(crud.ViewStyleListOnly).
		Title("Nutzerkonten").
		FindAll(service.AllUsers(subject))

	return crud.View[iam.User](opts)

}

type createUser struct {
	Firstname string
	Lastname  string
	EMail     string
	Password1 string
	Password2 string
}

func (createUser) Identity() string {
	return ""
}
