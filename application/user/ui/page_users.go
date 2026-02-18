// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiuser

import (
	"strings"

	"github.com/worldiety/i18n"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/consent"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/pkg/xstrings"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/avatar"
	"go.wdy.de/nago/presentation/ui/dataview"
	"go.wdy.de/nago/presentation/ui/form"
	"go.wdy.de/nago/presentation/ui/pager"
	"golang.org/x/text/language"
)

var (
	StrAccountTitle      = i18n.MustString("nago.admin.user.title_accounts", i18n.Values{language.English: "User Accounts", language.German: "Nutzerkonten"})
	StrNotifyUser        = i18n.MustString("nago.admin.user.notify", i18n.Values{language.English: "Notify user about account", language.German: "Nutzer über Konto benachrichtigen"})
	StrNotifyUserDesc    = i18n.MustString("nago.admin.user.notify_desc", i18n.Values{language.English: "Notify the user by email that this account has been created and ask them to log in. To do this, the same internal domain event is generated as if this user had been newly created. Processes or procedures that assume this event is unique may behave incorrectly.", language.German: "Den Nutzer per E-Mail darüber benachrichtigen, dass dieses Konto angelegt wurde und ihn auffordern sich anzumelden. Dazu wird das gleiche interne Domänen-Ereignis erzeugt, als ob dieser Nutzer neu angelegt wurde. Prozesse oder Abläufe die davon ausgehen, dass dieses Ereignis einmalig ist, können sich womöglich fehlerhaft verhalten."})
	StrExportUserCSV     = i18n.MustString("nago.admin.user.export_csv", i18n.Values{language.English: "Export users as CSV", language.German: "Nutzer als CSV exportieren"})
	StrExportUserCSVDesc = i18n.MustString("nago.admin.user.export_csv_desc", i18n.Values{language.English: "Export the selected users with their personal data to a CSV file. Please note that using this data without the users' consent could be a violation of the GDPR.", language.German: "Die ausgewählten Nutzer mit ihren persönlichen Daten in einer CSV Datei exportieren. Beachten Sie, dass die Verwendung dieser Daten ohne Zustimmung der Nutzer ein Verstoß gegen die DSGVO sein könnte."})
)

type UserModel struct {
	ID                    user.ID
	Email                 user.Email
	Contact               user.Contact
	EMailVerified         bool
	Status                user.AccountStatus
	RequirePasswordChange bool
	NLSManagedUser        bool
	VerificationCode      user.Code

	// some legal/regulatory properties
	Consents []consent.Consent `json:"consents,omitzero"`

	Roles             []role.ID
	Groups            []group.ID
	GlobalPermissions []permission.ID
}

func (u UserModel) IntoUser() user.User {
	return user.User{
		ID:                    u.ID,
		Email:                 u.Email,
		Contact:               u.Contact,
		EMailVerified:         u.EMailVerified,
		Status:                u.Status,
		RequirePasswordChange: u.RequirePasswordChange,
		NLSManagedUser:        u.NLSManagedUser,
		VerificationCode:      u.VerificationCode,
		Consents:              u.Consents,
	}
}

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

	editUserPresented := core.AutoState[bool](wnd)
	createUserPresented := core.AutoState[bool](wnd)
	selectedUser := core.AutoState[user.User](wnd)

	if editUserPresented.Get() && selectedUser.Get().ID != "" && model.Selections[selectedUser.Get().ID] == nil {
		selectedUser.Set(user.User{})
		editUserPresented.Set(false)
	}

	return ui.VStack(
		ui.H1(StrAccountTitle.Get(wnd)),

		dataview.FromData(wnd, dataview.Data[user.User, user.ID]{
			FindAll: ucUsers.FindAllIdentifiers(wnd.Subject()),
			FindByID: func(id user.ID) (option.Opt[user.User], error) {
				return ucUsers.FindByID(wnd.Subject(), id)
			},
			Fields: []dataview.Field[user.User]{
				{
					ID:   "avatar",
					Name: "",
					Map: func(obj user.User) core.View {
						display := xstrings.Join2(" ", obj.Contact.Firstname, obj.Contact.Lastname)
						return avatar.TextOrImage(display, obj.Contact.Avatar)
					},
				},

				{
					ID:   "name",
					Name: rstring.LabelName.Get(wnd),
					Map: func(obj user.User) core.View {
						display := xstrings.Join2(" ", obj.Contact.Firstname, obj.Contact.Lastname)
						return ui.Text(display)
					},
					Comparator: func(a, b user.User) int {
						return xstrings.CompareIgnoreCase(a.Contact.Firstname, b.Contact.Firstname)
					},
				},

				{
					ID:   "email",
					Name: rstring.LabelEMail.Get(wnd),
					Map: func(obj user.User) core.View {
						return ui.Text(string(obj.Email))
					},
					Comparator: func(a, b user.User) int {
						return xstrings.CompareIgnoreCase(string(a.Email), string(b.Email))
					},
				},

				{
					ID:   "status",
					Name: rstring.LabelState.Get(wnd),
					Map: func(obj user.User) core.View {
						return ui.Text(stateStr(obj))
					},
					Comparator: func(a, b user.User) int {
						return strings.Compare(stateStr(a), stateStr(b))
					},
				},
			},
		}).
			Search(true).
			CreateAction(func() {
				createUserPresented.Set(true)
			}).
			CardOptions(dataview.CardOptions{
				Title: "name",
				Hints: map[dataview.FieldID]dataview.FormatHint{
					"avatar": dataview.HintInvisible,
					"status": dataview.HintInline,
				},
			}).
			SelectOptions(
				dataview.NewSelectOptionDelete(wnd, func(selected []user.ID) error {
					for _, id := range selected {
						if err := ucUsers.Delete(wnd.Subject(), id); err != nil {
							return err
						}
					}

					return nil
				}),

				dataview.SelectOption[user.ID]{
					Icon: icons.MailBox,
					Name: StrNotifyUser.Get(wnd),
					Action: func(selected []user.ID) error {
						for _, id := range selected {
							optUsr, err := ucUsers.FindByID(wnd.Subject(), id)
							if err != nil {
								return err
							}
							if optUsr.IsNone() {
								// stale ref
								continue
							}

							usr := optUsr.Unwrap()
							user.PublishUserCreated(wnd.Application().EventBus(), usr, true)
						}

						return nil
					},
					ConfirmDialog: func(selected []user.ID) dataview.ConfirmDialog[user.ID] {
						return dataview.ConfirmDialog[user.ID]{
							Title:   StrNotifyUser.Get(wnd),
							Message: StrNotifyUserDesc.Get(wnd),
						}
					},
				},

				dataview.SelectOption[user.ID]{
					Icon: icons.MailBox,
					Name: StrExportUserCSV.Get(wnd),
					Action: func(selected []user.ID) error {
						buf, err := ucUsers.ExportUsers(wnd.Subject(), selected, user.ExportUsersOptions{
							Format:   user.ExportCSV,
							Language: wnd.Locale(),
						})
						if err != nil {
							return err
						}

						wnd.ExportFiles(core.ExportFileBytes("users.csv", buf))

						return nil
					},
					ConfirmDialog: func(selected []user.ID) dataview.ConfirmDialog[user.ID] {
						return dataview.ConfirmDialog[user.ID]{
							Title:   StrExportUserCSV.Get(wnd),
							Message: StrExportUserCSVDesc.Get(wnd),
						}
					},
				},
			).
			Action(func(u user.User) {
				selectedUser.Set(u)
				editUserPresented.Set(true)
			}).
			NextActionIndicator(true).
			TableOptions(dataview.TableOptions{
				ColumnWidths: map[dataview.FieldID]ui.Length{"avatar": ui.L64},
			}),

		ui.Space(ui.L64),
		dlgEditUser(wnd, ucUsers, ucGroups, ucRoles, ucPermissions, editUserPresented, selectedUser),
		dlgCreateUserModel(wnd, ucUsers, createUserPresented),
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

	transientUserClone := core.AutoState[UserModel](wnd).Init(func() UserModel {
		optUsr, err := ucUsers.FindByID(wnd.Subject(), selectedUsr.Get().ID)
		if err != nil {
			alert.ShowBannerError(wnd, err)
			return UserModel{}
		}

		usr := optUsr.UnwrapOr(user.User{})

		usrM := UserModel{
			ID:                    usr.ID,
			Email:                 usr.Email,
			Contact:               usr.Contact,
			EMailVerified:         usr.EMailVerified,
			Status:                usr.Status,
			RequirePasswordChange: usr.RequirePasswordChange,
			NLSManagedUser:        usr.NLSManagedUser,
			Consents:              usr.Consents,
			VerificationCode:      usr.VerificationCode,
		}

		roles, err := xslices.Collect2(ucUsers.ListRoles(wnd.Subject(), selectedUsr.Get().ID))
		if err != nil {
			alert.ShowBannerError(wnd, err)
			return UserModel{}
		}

		usrM.Roles = roles

		groups, err := xslices.Collect2(ucUsers.ListGroups(wnd.Subject(), selectedUsr.Get().ID))
		if err != nil {
			alert.ShowBannerError(wnd, err)
			return UserModel{}
		}

		usrM.Groups = groups

		perms, err := xslices.Collect2(ucUsers.ListGlobalPermissions(wnd.Subject(), selectedUsr.Get().ID))
		if err != nil {
			alert.ShowBannerError(wnd, err)
			return UserModel{}
		}
		usrM.GlobalPermissions = perms

		return usrM
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

		if err := ucUsers.UpdateOtherPermissions(wnd.Subject(), usr.ID, usr.GlobalPermissions); err != nil {
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
