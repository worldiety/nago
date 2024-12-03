package iamui

import (
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/presentation/core"
	heroOutline "go.wdy.de/nago/presentation/icons/hero/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/avatar"
	"go.wdy.de/nago/presentation/ui/crud"
	"go.wdy.de/nago/presentation/ui/list"
)

func Users(wnd core.Window, service *iam.Service) core.View {
	subject := wnd.Subject()
	if !subject.Valid() {
		return alert.BannerError(auth.NotLoggedIn(""))
	}

	bnd := crud.NewBinding[iam.User](wnd).
		DeleteFunc(func(e iam.User) error {
			return service.DeleteUser(subject, e.ID)
		}).
		EntityName("Nutzer")
	bnd.Add(
		//crud.Text(crud.TextOptions{Label: "ID"}, crud.Ptr(func(e *iam.User) *auth.UID {
		//	return &e.ID
		//})).ReadOnly(true).WithoutTable(),
		crud.Text(crud.TextOptions{Label: "Vorname"}, crud.Ptr(func(e *iam.User) *string {
			return &e.Firstname
		})),
		crud.Text(crud.TextOptions{Label: "Nachname"}, crud.Ptr(func(e *iam.User) *string {
			return &e.Lastname
		})),
		//crud.OneToMany(crud.OneToManyOptions[iam.Permission, iam.PID]{
		//	Label:           "Vererbte Berechtigungen",
		//	ForeignEntities: service.AllPermissions(subject),
		//	ForeignPickerRenderer: func(t iam.Permission) core.View {
		//		return ui.Text(t.Name())
		//	},
		//}, crud.Ptr(func(user *iam.User) *[]iam.PID {
		//	// this is by intention: we show the entire inherited list of all permission, not just the customized ones
		//	perms, err := service.FindAllUserPermissions(subject, user.ID)
		//	if err != nil {
		//		alert.ShowBannerError(wnd, err)
		//	}
		//
		//	return &perms
		//})).ReadOnly(true),

		crud.OneToMany(crud.OneToManyOptions[iam.Role, auth.RID]{
			Label:           "Rollen",
			ForeignEntities: service.AllRoles(subject),
			ForeignPickerRenderer: func(t iam.Role) core.View {
				return ui.Text(t.Name)
			},
		}, crud.Ptr(func(model *iam.User) *[]auth.RID {
			return &model.Roles
		})),

		crud.OneToMany(crud.OneToManyOptions[iam.Permission, iam.PID]{
			Label:           "Einzelberechtigungen",
			SupportingText:  "Diese Berechtigungen sollten nur in Ausnahmefällen vergeben werden und ansonsten über Rollen abgebildet werden.",
			ForeignEntities: service.AllPermissions(subject),
			ForeignPickerRenderer: func(t iam.Permission) core.View {
				return ui.Text(t.Name())
			},
		}, crud.Ptr(func(user *iam.User) *[]iam.PID {
			return &user.Permissions
		})),

		crud.OneToMany(crud.OneToManyOptions[iam.Group, auth.GID]{
			Label:           "Gruppen",
			SupportingText:  "Die Gruppenzugehörigkeit gehört zu den ressourcenbasierten Berechtigungen.",
			ForeignEntities: service.AllGroups(subject),
			ForeignPickerRenderer: func(t iam.Group) core.View {
				return ui.Text(t.Name)
			},
		}, crud.Ptr(func(model *iam.User) *[]auth.GID {
			return &model.Groups
		})),

		//crud.OneToMany(crud.OneToManyOptions[license.License, license.ID]{
		//	Label:           "Nutzerbasierte Lizenzen",
		//	ForeignEntities: license.Global.All(),
		//	ForeignPickerRenderer: func(t license.License) core.View {
		//		return ui.Text(t.LicenseName())
		//	},
		//}, crud.Ptr(func(model *iam.User) *[]license.ID {
		//	return &model.Licenses
		//})),

		//crud.AggregateActions(
		//	"Optionen",
		//	crud.ButtonDelete(wnd, func(group iam.User) error {
		//		return service.DeleteUser(subject, group.ID)
		//	}),
		//	crud.ButtonEdit(bnd, func(user iam.User) (errorText string, infrastructureError error) {
		//		return "", service.UpdateUser(subject, user.ID, string(user.Email), user.Firstname, user.Lastname, user.Permissions, user.Roles, user.Groups)
		//	}),
		//),
	).IntoListEntry(func(entity iam.User) list.TEntry {
		perms, err := service.FindAllUserPermissions(subject, entity.ID)
		if err != nil {
			alert.ShowBannerError(wnd, err)
		}
		editPresented := core.StateOf[bool](wnd, "crud.user.list.update-"+string(entity.ID))

		return list.Entry().
			Leading(avatar.Text(entity.Firstname + " " + entity.Lastname)).
			Trailing(ui.HStack(
				ui.ImageIcon(heroOutline.ChevronRight),
				crud.DialogEdit(bnd, editPresented, entity, func(user iam.User) (errorText string, infrastructureError error) {
					return "", service.UpdateUser(subject, user.ID, string(user.Email), user.Firstname, user.Lastname, user.Permissions, user.Roles, user.Groups)
				}),
			)).
			Action(func() {
				editPresented.Set(true)
			}).
			Headline(entity.Firstname + " " + entity.Lastname).
			SupportingText(fmt.Sprintf("%d Berechtigungen, %d Rollen", len(perms), len(entity.Roles)))
	})

	bndCrUsr := crud.NewBinding[createUser](wnd).EntityName("Nutzer").Add(
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
			_, err := service.NewUser(subject, model.EMail, model.Firstname, model.Lastname, iam.Password(model.Password1), iam.Password(model.Password2))
			return "", err
		})).
		ViewStyle(crud.ViewStyleListOnly).
		Title("Nutzerkonten").
		FindAll(service.AllUsers(subject))

	return crud.View[iam.User](opts).Frame(ui.Frame{}.FullWidth())

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
