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
		crud.Text("ID", func(e *iam.User) *auth.UID {
			return &e.ID
		}).ReadOnly(true).WithoutTable(),
		crud.Text("Vorname", func(e *iam.User) *string {
			return &e.Firstname
		}),
		crud.Text("Nachname", func(e *iam.User) *string {
			return &e.Lastname
		}),
		crud.OneToMany("Vererbte Berechtigungen", service.AllPermissions(subject), func(t iam.Permission) core.View {
			return ui.Text(t.Name())
		}, func(user *iam.User) *[]iam.PID {
			// this is by intention: we show the entire inherited list of all permission, not just the customized ones
			perms, err := service.FindAllUserPermissions(subject, user.ID)
			if err != nil {
				slog.Error("failed to find all user permissions", "err", err, "uid", user.ID)
			}

			return &perms
		}).ReadOnly(true),

		crud.OneToMany("Einzelberechtigungen", service.AllPermissions(subject), func(t iam.Permission) core.View {
			return ui.Text(t.Name())
		}, func(user *iam.User) *[]iam.PID {
			return &user.Permissions
		}),

		crud.OneToMany("Gruppen", service.AllGroups(subject), func(t iam.Group) core.View {
			return ui.Text(t.Name)
		}, func(model *iam.User) *[]auth.GID {
			return &model.Groups
		}),

		crud.OneToMany("Rollen", service.AllRoles(subject), func(t iam.Role) core.View {
			return ui.Text(t.Name)
		}, func(model *iam.User) *[]auth.RID {
			return &model.Roles
		}),

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
		crud.Text("Vorname", func(e *createUser) *string {
			return &e.Firstname
		}),
		crud.Text("Nachname", func(e *createUser) *string {
			return &e.Lastname
		}),
		crud.Text("eMail", func(e *createUser) *string {
			return &e.EMail
		}),
		crud.Password("Kennwort", func(e *createUser) *string {
			return &e.Password1
		}),
		crud.Password("Kennwort wiederholen", func(e *createUser) *string {
			return &e.Password2
		}),
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
		})).Title("Nutzerkonten").
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
