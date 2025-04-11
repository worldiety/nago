// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiuser

import (
	"fmt"
	"go.wdy.de/nago/application/billing"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	heroOutline "go.wdy.de/nago/presentation/icons/hero/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/avatar"
	"go.wdy.de/nago/presentation/ui/crud"
	"go.wdy.de/nago/presentation/ui/list"
	"golang.org/x/text/language"
)

type Pages struct {
	Users         core.NavigationPath
	MyProfile     core.NavigationPath
	MyContact     core.NavigationPath
	ConfirmMail   core.NavigationPath
	ResetPassword core.NavigationPath
	Register      core.NavigationPath
}

func Users(
	wnd core.Window,
	deleteUser user.Delete,
	allUsers user.FindAll,
	createUser user.Create,
	updateContact user.UpdateOtherContact,
	updateGroups user.UpdateOtherGroups,
	updateRoles user.UpdateOtherRoles,
	updatePermissions user.UpdateOtherPermissions,
	updateLicenses user.UpdateOtherLicenses,
	allRoles role.FindAll,
	allPermissions permission.FindAll,
	allGroups group.FindAll,
	viewOf user.SubjectFromUser,
	userLicenses billing.UserLicenses,
) core.View {
	subject := wnd.Subject()
	if !subject.Valid() {
		return alert.BannerError(session.NotLoggedInErr)
	}

	stats, err := userLicenses(subject)
	if err != nil {
		return alert.BannerError(err)
	}

	bnd := crud.NewBinding[user.User](wnd).
		DeleteFunc(func(e user.User) error {
			return deleteUser(subject, e.ID)
		}).
		EntityName("Nutzer")
	bnd.Add(
		//crud.Text(crud.TextOptions{Label: "EID"}, crud.Ptr(func(e *iam.User) *auth.UID {
		//	return &e.EID
		//})).ReadOnly(true).WithoutTable(),
		crud.Text(crud.TextOptions{Label: "Vorname"}, crud.Ptr(func(e *user.User) *string {
			return &e.Contact.Firstname
		})),
		crud.Text(crud.TextOptions{Label: "Nachname"}, crud.Ptr(func(e *user.User) *string {
			return &e.Contact.Lastname
		})),

		crud.OneToMany(crud.OneToManyOptions[role.Role, role.ID]{
			Label:           "Rollen",
			ForeignEntities: allRoles(subject),
			ForeignPickerRenderer: func(t role.Role) core.View {
				return ui.Text(t.Name)
			},
		}, crud.Ptr(func(model *user.User) *[]role.ID {
			return &model.Roles
		})),

		crud.OneToMany(crud.OneToManyOptions[permission.Permission, permission.ID]{
			Label:           "Einzelberechtigungen",
			SupportingText:  "Diese Berechtigungen sollten nur in Ausnahmefällen vergeben werden und ansonsten über Rollen abgebildet werden.",
			ForeignEntities: allPermissions(subject),
			ForeignPickerRenderer: func(t permission.Permission) core.View {
				return ui.Text(t.Name)
			},
		}, crud.Ptr(func(user *user.User) *[]permission.ID {
			return &user.Permissions
		})),

		crud.OneToMany(crud.OneToManyOptions[group.Group, group.ID]{
			Label:           "Gruppen",
			SupportingText:  "Die Gruppenzugehörigkeit gehört zu den ressourcenbasierten Berechtigungen.",
			ForeignEntities: allGroups(subject),
			ForeignPickerRenderer: func(t group.Group) core.View {
				return ui.Text(t.Name)
			},
		}, crud.Ptr(func(model *user.User) *[]group.ID {
			return &model.Groups
		})),

		crud.CustomView(func(usr *core.State[user.User]) ui.DecoredView {
			return licensePicker(wnd, stats, usr)
		}),
	).IntoListEntry(func(entity user.User) list.TEntry {
		optView, err := viewOf(subject, entity.ID)
		if err != nil {
			alert.ShowBannerError(wnd, err)
			return list.Entry()
		}

		if optView.IsNone() {
			return list.Entry()
		}

		view := optView.Unwrap()
		permCount := 0
		for range view.Permissions() {
			permCount++
		}

		editPresented := core.StateOf[bool](wnd, "crud.user.list.update-"+string(entity.ID))

		return list.Entry().
			Leading(avatar.Text(view.Name())).
			Trailing(ui.HStack(
				ui.ImageIcon(heroOutline.ChevronRight),
				crud.DialogEdit(bnd, editPresented, entity, func(user user.User) (errorText string, infrastructureError error) {
					if err := updateContact(wnd.Subject(), user.ID, user.Contact); err != nil {
						return "", err
					}

					if err := updatePermissions(wnd.Subject(), user.ID, user.Permissions); err != nil {
						return "", err
					}

					if err := updateRoles(wnd.Subject(), user.ID, user.Roles); err != nil {
						return "", err
					}

					if err := updateGroups(wnd.Subject(), user.ID, user.Groups); err != nil {
						return "", err
					}

					if err := updateLicenses(wnd.Subject(), user.ID, user.Licenses); err != nil {
						return "", err
					}

					return "", nil
				}),
			)).
			Action(func() {
				editPresented.Set(true)
			}).
			Headline(view.Name()).
			SupportingView(ui.Text(view.Email())).
			SupportingText(fmt.Sprintf("%d Berechtigungen, %d Rollen", permCount, len(entity.Roles)))
	})

	bndCrUsr := crud.NewBinding[createUserModel](wnd).EntityName("Nutzer").Add(
		crud.Text(crud.TextOptions{Label: "Vorname"}, crud.Ptr(func(e *createUserModel) *string {
			return &e.Firstname
		})),
		crud.Text(crud.TextOptions{Label: "Nachname"}, crud.Ptr(func(e *createUserModel) *string {
			return &e.Lastname
		})),
		crud.Text(crud.TextOptions{Label: "eMail"}, crud.Ptr(func(e *createUserModel) *string {
			return &e.EMail
		})),
		crud.Password(crud.PasswordOptions{Label: "Kennwort", SupportingText: "Bleibt das Kennwort leer, muss der Nutzer bei der ersten Anmeldung sein Kennwort ändern."}, crud.Ptr(func(e *createUserModel) *string {
			return &e.Password1
		})),
		crud.Password(crud.PasswordOptions{Label: "Kennwort wiederholen"}, crud.Ptr(func(e *createUserModel) *string {
			return &e.Password2
		})),
		crud.Bool(crud.BoolOptions{Label: "Nutzer verifiziert"}, crud.Ptr(func(e *createUserModel) *bool {
			return &e.Verified
		})),
		crud.Bool(crud.BoolOptions{Label: "Nutzer benachrichtigen"}, crud.Ptr(func(e *createUserModel) *bool {
			return &e.Notify
		})),
	)

	opts := crud.Options(bnd).
		Actions(crud.ButtonCreate[createUserModel](bndCrUsr, createUserModel{}, func(model createUserModel) (errorText string, infrastructureError error) {
			_, err := createUser(subject, user.ShortRegistrationUser{
				Firstname:         model.Firstname,
				Lastname:          model.Lastname,
				Email:             user.Email(model.EMail),
				Password:          user.Password(model.Password1),
				PasswordRepeated:  user.Password(model.Password2),
				PreferredLanguage: language.German,
				NotifyUser:        model.Notify,
				Verified:          model.Verified,
			})
			return "", err
		})).
		ViewStyle(crud.ViewStyleListOnly).
		Title("Nutzerkonten").
		FindAll(allUsers(subject))

	return crud.View[user.User](opts).Frame(ui.Frame{}.FullWidth())

}

type createUserModel struct {
	Firstname string
	Lastname  string
	EMail     string
	Password1 string
	Password2 string
	Notify    bool
	Verified  bool
}

func (createUserModel) Identity() string {
	return ""
}
