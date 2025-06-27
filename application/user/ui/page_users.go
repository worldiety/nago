// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiuser

import (
	"fmt"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/data/rquery"
	"go.wdy.de/nago/pkg/xstrings"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/avatar"
	"go.wdy.de/nago/presentation/ui/form"
	"go.wdy.de/nago/presentation/ui/pager"
	"golang.org/x/text/language"
)

func PageUsers(wnd core.Window, ucUsers user.UseCases, ucGroups group.UseCases, ucRoles role.UseCases, ucPermissions permission.UseCases) core.View {
	if err := wnd.Subject().Audit(user.PermFindAll); err != nil {
		return alert.BannerError(err)
	}

	query := core.AutoState[string](wnd)
	pageIdx := core.AutoState[int](wnd)

	filterOpts := data.FilterOptions[user.User, user.ID]{}
	if query.Get() != "" {
		filterOpts.Accept = rquery.SimplePredicate[user.User](query.Get())
	}

	page, err := data.FilterAndPaginate[user.User, user.ID](
		func(id user.ID) (option.Opt[user.User], error) {
			return ucUsers.FindByID(wnd.Subject(), id)
		},
		ucUsers.FindAllIdentifiers(wnd.Subject()),
		filterOpts,
		data.PaginateOptions{
			PageIdx: pageIdx.Get(),
		},
	)

	if err != nil {
		return alert.BannerError(err)
	}

	editUserPresented := core.AutoState[bool](wnd)
	createUserPresented := core.AutoState[bool](wnd)
	selectedUser := core.AutoState[user.User](wnd)

	allSelected := core.AutoState[bool](wnd)
	checkboxModel := core.AutoState[map[user.ID]*core.State[bool]](wnd).Init(func() map[user.ID]*core.State[bool] {
		return map[user.ID]*core.State[bool]{}
	})

	return ui.VStack(
		ui.H1("Nutzerkonten"),
		dlgEditUser(wnd, ucUsers, ucGroups, ucRoles, ucPermissions, editUserPresented, selectedUser),
		dlgCreateUserModel(wnd, ucUsers, createUserPresented),
		ui.HStack(
			ui.TextField("", query.Get()).InputValue(query).Style(ui.TextFieldReduced).Leading(ui.ImageIcon(icons.Search)),
			ui.PrimaryButton(func() {
				createUserPresented.Set(true)
			}).Title("Nutzer hinzufügen"),
		).Alignment(ui.Trailing).FullWidth().Gap(ui.L8),
		ui.Space(ui.L32),

		ui.Table(
			ui.TableColumn(ui.Checkbox(allSelected.Get()).InputChecked(allSelected)).Width(ui.L64),
			ui.TableColumn(ui.Text("")).Width(ui.L64),
			ui.TableColumn(ui.Text("Anzeigename")),
			ui.TableColumn(ui.Text("E-Mail")),
			ui.TableColumn(ui.Text("Status")),
			ui.TableColumn(ui.Text("")).Width(ui.L64),
		).Rows(
			ui.ForEach(page.Items, func(u user.User) ui.TTableRow {
				myState := core.StateOf[bool](wnd, "checkbox-"+string(u.ID))
				checkboxModel.Get()[u.ID] = myState

				display := xstrings.Join2(" ", u.Contact.Firstname, u.Contact.Lastname)
				return ui.TableRow(
					ui.TableCell(ui.Checkbox(false)),
					ui.TableCell(avatar.TextOrImage(display, u.Contact.Avatar)),
					ui.TableCell(ui.Text(display)),
					ui.TableCell(ui.Text(string(u.Email))),
					ui.TableCell(ui.Text(stateStr(u))),
					ui.TableCell(ui.ImageIcon(icons.ChevronRight)),
				).Action(func() {
					selectedUser.Set(u)
					editUserPresented.Set(true)
				}).HoveredBackgroundColor(ui.ColorCardFooter)
			})...,
		).Rows(
			ui.TableRow(
				ui.TableCell(
					ui.HStack(
						ui.Text(fmt.Sprintf("%d-%d von %d Einträgen", page.PageIdx*page.PageSize+1, page.PageIdx*page.PageSize+page.PageSize, page.Total)),
						ui.Spacer(),
						pager.Pager(pageIdx).Count(page.PageCount),
					).FullWidth(),
				).ColSpan(6),
			).BackgroundColor(ui.ColorCardFooter),
		).
			Frame(ui.Frame{}.FullWidth()),
	).FullWidth().Alignment(ui.Leading)

}

func stateStr(usr user.User) string {
	if usr.Enabled() {
		if usr.RequiresVerification() {
			return "Verifizierung erforderlich"
		}

		return "Aktiv"
	} else {
		return "Deaktiviert"
	}
}

func dlgEditUser(wnd core.Window, ucUsers user.UseCases, ucGroups group.UseCases, ucRoles role.UseCases, ucPermissions permission.UseCases, presented *core.State[bool], selectedUsr *core.State[user.User]) core.View {
	if !presented.Get() {
		return nil
	}

	transientUserClone := core.AutoState[user.User](wnd).Init(func() user.User {
		optUsr, err := ucUsers.FindByID(wnd.Subject(), selectedUsr.Get().ID)
		if err != nil {
			alert.ShowBannerError(wnd, err)
			return user.User{}
		}

		return optUsr.UnwrapOr(user.User{})
	})

	return alert.Dialog("Nutzer bearbeiten", ViewEditUser(wnd, ucUsers, ucGroups, ucRoles, ucPermissions, transientUserClone).Frame(ui.Frame{Height: ui.Full, Width: ui.Full}), presented, alert.Closeable(), alert.XLarge(), alert.FullHeight(), alert.Cancel(nil), alert.Save(func() (close bool) {
		usr := transientUserClone.Get()
		if err := ucUsers.UpdateOtherContact(wnd.Subject(), usr.ID, usr.Contact); err != nil {
			alert.ShowBannerError(wnd, err)
			return false
		}

		if err := ucUsers.UpdateOtherRoles(wnd.Subject(), usr.ID, usr.Roles); err != nil {
			alert.ShowBannerError(wnd, err)
			return false
		}

		if err := ucUsers.UpdateOtherGroups(wnd.Subject(), usr.ID, usr.Groups); err != nil {
			alert.ShowBannerError(wnd, err)
			return false
		}

		if err := ucUsers.UpdateOtherPermissions(wnd.Subject(), usr.ID, usr.Permissions); err != nil {
			alert.ShowBannerError(wnd, err)
			return false
		}

		return true
	}))
}

type createUserModel struct {
	Firstname string `label:"Vorname"`
	Lastname  string `label:"Nachname"`
	EMail     string `label:"E-Mail"`
	Password1 string `label:"Kennwort" supportingText:"Das Kennwort kann leer bleiben, muss dann aber per Double-OptIn vom Nutzer selbst gesetzt werden." style:"secret"`
	Password2 string `label:"Kennwort wiederholen" style:"secret"`
	Notify    bool   `label:"Nutzer benachrichtigen"`
	Verified  bool   `label:"Nutzer als bereits verifiziert markieren"`
}

func (m createUserModel) IntoShortRegistration() user.ShortRegistrationUser {
	return user.ShortRegistrationUser{
		Firstname:         m.Firstname,
		Lastname:          m.Lastname,
		Email:             user.Email(m.EMail),
		Password:          user.Password(m.Password1),
		PasswordRepeated:  user.Password(m.Password2),
		PreferredLanguage: language.German,
		NotifyUser:        m.Notify,
		Verified:          m.Verified,
	}
}

func dlgCreateUserModel(wnd core.Window, ucUsers user.UseCases, presented *core.State[bool]) core.View {
	if !presented.Get() {
		return nil
	}

	model := core.AutoState[createUserModel](wnd)
	errState := core.AutoState[error](wnd)

	strength := user.CalculatePasswordStrength(model.Get().Password1)
	return alert.Dialog(
		"Nutzer erstellen",
		ui.VStack(
			ui.Form(
				form.Auto(form.AutoOptions{Window: wnd}, model),
			).Autocomplete(false).Frame(ui.Frame{Width: ui.Full}),
			ui.Space(ui.L32),
			PasswordStrengthView(wnd, strength),
			ui.If(errState.Get() != nil, alert.BannerError(errState.Get())),
		).FullWidth(),
		presented,
		alert.Closeable(),
		alert.Large(),
		alert.Cancel(nil),
		alert.Save(func() (close bool) {
			if _, err := ucUsers.Create(wnd.Subject(), model.Get().IntoShortRegistration()); err != nil {
				errState.Set(err)
				return false
			}

			return true
		}),
	)

}
