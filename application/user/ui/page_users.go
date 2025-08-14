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

	model, err := pager.NewModel(
		wnd,
		func(id user.ID) (option.Opt[user.User], error) {
			return ucUsers.FindByID(wnd.Subject(), id)
		},
		ucUsers.FindAllIdentifiers(wnd.Subject()),
		pager.ModelOptions{},
	)

	if err != nil {
		return alert.BannerError(err)
	}

	batchEtcPresented := core.AutoState[bool](wnd)
	editUserPresented := core.AutoState[bool](wnd)
	createUserPresented := core.AutoState[bool](wnd)
	selectedUser := core.AutoState[user.User](wnd)

	if editUserPresented.Get() && selectedUser.Get().ID != "" && model.Selections[selectedUser.Get().ID] == nil {
		selectedUser.Set(user.User{})
		editUserPresented.Set(false)
	}

	return ui.VStack(
		ui.H1("Nutzerkonten"),
		func() core.View {
			if !batchEtcPresented.Get() {
				return nil
			}

			idents := model.Selected()
			return dialogEtcBatch(wnd, ucUsers, batchEtcPresented, idents)
		}(),
		dlgEditUser(wnd, ucUsers, ucGroups, ucRoles, ucPermissions, editUserPresented, selectedUser),
		dlgCreateUserModel(wnd, ucUsers, createUserPresented),
		ui.HStack(
			ui.TextField("", model.Query.Get()).InputValue(model.Query).Style(ui.TextFieldReduced).Leading(ui.ImageIcon(icons.Search)),
			ui.TertiaryButton(func() {
				model.UnselectAll()
			}).Title(fmt.Sprintf("%d ausgewählt", model.SelectionCount)).PostIcon(icons.Close).Visible(model.SelectionCount > 0),
			ui.SecondaryButton(func() {
				batchEtcPresented.Set(true)
			}).Title("Optionen").PreIcon(icons.Grid).Enabled(model.SelectionCount > 0),
			ui.VLineWithColor(ui.ColorInputBorder).Frame(ui.Frame{Height: ui.L40}),
			ui.PrimaryButton(func() {
				createUserPresented.Set(true)
			}).Title("Nutzer hinzufügen"),
		).Alignment(ui.Trailing).FullWidth().Gap(ui.L8),
		ui.Space(ui.L32),

		ui.Table(
			ui.TableColumn(ui.Checkbox(model.SelectSubset.Get()).InputChecked(model.SelectSubset)).Width(ui.L64),
			ui.TableColumn(ui.Text("")).Width(ui.L64),
			ui.TableColumn(ui.Text("Anzeigename")),
			ui.TableColumn(ui.Text("E-Mail")),
			ui.TableColumn(ui.Text("Status")),
			ui.TableColumn(ui.Text("")).Width(ui.L64),
		).Rows(
			ui.ForEach(model.Page.Items, func(u user.User) ui.TTableRow {
				myState := model.Selections[u.ID]

				display := xstrings.Join2(" ", u.Contact.Firstname, u.Contact.Lastname)
				return ui.TableRow(
					ui.TableCell(ui.Checkbox(myState.Get()).InputChecked(myState)),
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
						ui.Text(model.PageString()),
						ui.Spacer(),
						pager.Pager(model.PageIdx).Count(model.Page.PageCount),
					).FullWidth(),
				).ColSpan(6),
			).BackgroundColor(ui.ColorCardFooter),
		).
			Frame(ui.Frame{}.FullWidth()),
	).FullWidth().Alignment(ui.Leading)

}

func stateStr(usr user.User) string {
	if usr.Enabled() {
		if !usr.VerificationCode.IsZero() {
			return "E-Mail-Bestätigung erforderlich"
		}

		if usr.RequirePasswordChange {
			return "Kennwort erforderlich"
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
