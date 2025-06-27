// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiuser

import (
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/form"
	"go.wdy.de/nago/presentation/ui/list"
	"go.wdy.de/nago/presentation/ui/tabs"
	"time"
)

func ViewEditUser(wnd core.Window, ucUsers user.UseCases, ucGroups group.UseCases, ucRoles role.UseCases, ucPermissions permission.UseCases, usr *core.State[user.User]) ui.DecoredView {
	editContact := core.AutoState[contactViewModel](wnd).Init(func() contactViewModel {
		return newContactViewModel(string(usr.Get().Email), usr.Get().Contact)
	}).Observe(func(c contactViewModel) {
		u := usr.Get()
		u.Contact = c.IntoContact(wnd)
		usr.Set(u)
		usr.Notify()
	})

	pageIdx := core.AutoState[int](wnd)
	return ui.VStack(
		tabs.Tabs(
			tabs.Page("Kontakt", func() core.View {
				return viewContact(wnd, editContact)
			}).Icon(icons.AddressBook),
			tabs.Page("Rollen", func() core.View {
				return viewRoles(wnd, ucUsers, ucRoles, usr)
			}).Icon(icons.UserSettings),
			tabs.Page("Gruppen", func() core.View {
				return viewGroups(wnd, ucGroups, usr)
			}).Icon(icons.UsersGroup),
			tabs.Page("Berechtigungen", func() core.View {
				return viewPermissions(wnd, usr)
			}).Icon(icons.Shield),
			tabs.Page("Ressourcen", func() core.View {
				return ui.Text("TODO")
			}).Icon(icons.Shield).Disabled(true),
			tabs.Page("Zustimmungen", func() core.View {
				return viewConsents(wnd, ucUsers, usr)
			}).Icon(icons.Bookmark),
			tabs.Page("Lizenzen", func() core.View {
				return ui.Text("todo")
			}).Icon(icons.Book).Disabled(true),
			tabs.Page("Sonstiges", func() core.View {
				return viewEtc(wnd, ucUsers, usr)
			}).Icon(icons.UserSettings),
		).InputValue(pageIdx).Frame(ui.Frame{}.FullWidth()),
	).FullWidth().Alignment(ui.Leading)
}

func viewRoles(wnd core.Window, ucUsers user.UseCases, ucRoles role.UseCases, usr *core.State[user.User]) core.View {
	type viewModelRoles struct {
		Roles []role.ID `label:"Rollen" source:"nago.roles"`
	}

	editRoles := core.AutoState[viewModelRoles](wnd).Init(func() viewModelRoles {
		return viewModelRoles{
			Roles: usr.Get().Roles,
		}
	}).Observe(func(c viewModelRoles) {
		u := usr.Get()
		u.Roles = c.Roles
		usr.Set(u)
		usr.Notify()
	})

	var rolesView []core.View
	for _, id := range usr.Get().Roles {
		var permsViewInRole []core.View
		optRole, err := ucRoles.FindByID(wnd.Subject(), id)
		if err != nil {
			return alert.BannerError(err)
		}

		if optRole.IsZero() {
			// ignore stale ref
			continue
		}

		r := optRole.Unwrap()
		for _, pid := range r.Permissions {
			if perm, ok := permission.Find(pid); ok {
				permsViewInRole = append(permsViewInRole, list.Entry().Leading(ui.ImageIcon(icons.Shield)).Headline(perm.Name).SupportingText(perm.Description))
			}
		}

		rolesView = append(rolesView, list.List(permsViewInRole...).Caption(ui.Text(r.Name)))
	}

	return ui.VStack(
		form.Auto(form.AutoOptions{Window: wnd}, editRoles).Frame(ui.Frame{Width: ui.Full}),
		ui.If(len(rolesView) > 0, ui.Text("Die selektierten Rollen vererben die folgenden Berechtigungen.")),
		ui.If(len(rolesView) == 0, ui.Text("Es sind keine Rollen ausgewählt, durch die Berechtigungen vererbt werden.")),
	).Append(rolesView...).FullWidth().Alignment(ui.Leading).Gap(ui.L32)
}

func viewGroups(wnd core.Window, ucGroups group.UseCases, usr *core.State[user.User]) core.View {
	type viewModelGroups struct {
		Groups []group.ID `label:"Gruppen" source:"nago.groups"`
	}

	editGroups := core.AutoState[viewModelGroups](wnd).Init(func() viewModelGroups {
		return viewModelGroups{
			Groups: usr.Get().Groups,
		}
	}).Observe(func(c viewModelGroups) {
		u := usr.Get()
		u.Groups = c.Groups
		usr.Set(u)
		usr.Notify()
	})

	var groupsView []core.View

	for _, id := range usr.Get().Groups {
		optGroup, err := ucGroups.FindByID(wnd.Subject(), id)
		if err != nil {
			return alert.BannerError(err)
		}

		if optGroup.IsZero() {
			continue
		}

		grp := optGroup.Unwrap()
		groupsView = append(groupsView, list.Entry().Leading(ui.ImageIcon(icons.UsersGroup)).Headline(grp.Name).SupportingText(grp.Description))
	}

	if len(groupsView) == 0 {
		groupsView = append(groupsView, list.Entry().Headline("Keine Gruppenmitgliedschaften vorhanden."))
	}

	return ui.VStack(
		form.Auto(form.AutoOptions{Window: wnd}, editGroups).Frame(ui.Frame{Width: ui.Full}),
	).Append(list.List(groupsView...).Caption(ui.Text("Mitgliedschaften")).FullWidth()).FullWidth().Alignment(ui.Leading).Gap(ui.L32)

}

func viewContact(wnd core.Window, usr *core.State[contactViewModel]) core.View {
	return form.Auto(form.AutoOptions{Window: wnd}, usr).Frame(ui.Frame{Width: ui.Full})
}

func viewPermissions(wnd core.Window, usr *core.State[user.User]) core.View {
	type viewModelPerms struct {
		Perms []permission.ID `label:"Berechtigungen" source:"nago.permissions"`
	}

	editPerms := core.AutoState[viewModelPerms](wnd).Init(func() viewModelPerms {
		return viewModelPerms{
			Perms: usr.Get().Permissions,
		}
	}).Observe(func(c viewModelPerms) {
		u := usr.Get()
		u.Permissions = c.Perms
		usr.Set(u)
		usr.Notify()
	})

	var permsView []core.View
	for _, id := range usr.Get().Permissions {
		perm, ok := permission.Find(id)
		if !ok {
			// stale ref
			continue
		}

		permsView = append(permsView, list.Entry().Leading(ui.ImageIcon(icons.Shield)).Headline(perm.Name).SupportingText(perm.Description))
	}

	if len(permsView) == 0 {
		msg := "Keine einzeln vergebenen Berechtigungen vorhanden."
		if len(usr.Get().Roles) > 0 {
			msg += " Es gibt noch weitere Berechtigungen, die durch Rollen vererbt wurden."
		}
		permsView = append(permsView, list.Entry().Headline(msg))
	}

	return ui.VStack(
		ui.Text("Die hier einzeln ausgewählten Berechtigungen gelten nur für dieses Konto. Grundsätzlich ist es empfehlenswert Berechtigungen als Rolle zu modellieren, anstatt diese einzeln zu vergeben."),
		form.Auto(form.AutoOptions{Window: wnd}, editPerms).Frame(ui.Frame{Width: ui.Full}),
		ui.If(len(permsView) > 0 && len(usr.Get().Roles) > 0, ui.Text("Es gibt noch weitere Berechtigungen, die durch Rollen vererbt wurden.")),
	).Append(list.List(permsView...).Caption(ui.Text("Einzelberechtigungen")).FullWidth()).FullWidth().Alignment(ui.Leading).Gap(ui.L32)

}

func viewConsents(wnd core.Window, ucUser user.UseCases, usr *core.State[user.User]) core.View {

	return ui.VStack(
		ui.Text("Die Änderungen an den Zustimmungen werden sofort angewendet und können nicht durch 'Abbrechen' rückgängig gemacht werden."),
		actionCard(wnd, nil, usr.Get().ID, ucUser.FindByID, ucUser.Consent),
	).FullWidth().Alignment(ui.Leading).Gap(ui.L32)
}

func viewEtc(wnd core.Window, ucUsers user.UseCases, usr *core.State[user.User]) core.View {
	return ui.VStack(
		ui.VStack(
			ui.H2("Nutzer über Konto benachrichtigen"),
			ui.Text("Den Nutzer per E-Mail darüber benachrichtigen, dass dieses Konto angelegt wurde und ihn auffordern sich anzumelden. Dazu wird das gleiche interne Domänen-Ereignis erzeugt, als ob dieser Nutzer neu angelegt wurde. Prozesse oder Abläufe die davon ausgehen, dass dieses Ereignis einmalig ist, können sich womöglich fehlerhaft verhalten."),
			ui.HStack(
				ui.SecondaryButton(func() {
					bus := wnd.Application().EventBus()
					user.PublishUserCreated(bus, usr.Get(), true)
					alert.ShowBannerMessage(wnd, alert.Message{
						Title:    "Nutzer erstellt",
						Message:  "Ereignis für " + usr.String() + " erstellt.",
						Intent:   alert.IntentOk,
						Duration: time.Second * 2,
					})
				}).Title("Nutzer benachrichtigen"),
			).FullWidth().Alignment(ui.Trailing),
		).FullWidth().Alignment(ui.Leading).Gap(ui.L8).Border(ui.Border{}.Radius(ui.L16).Width(ui.L1).Color(ui.ColorInputBorder)).Padding(ui.Padding{}.All(ui.L16)),
	).FullWidth().Alignment(ui.Leading).Gap(ui.L32)
}
