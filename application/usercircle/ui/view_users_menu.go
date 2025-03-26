// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiusercircles

import (
	"fmt"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/application/usercircle"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
)

func makeMenu(wnd core.Window, circle usercircle.Circle, selectedCount int, userStates []userState, useCases usercircle.UseCases, removeUser func(users []user.User)) core.View {
	hasUserMenu := circle.CanEnable || circle.CanDisable || circle.CanDelete || circle.CanVerify

	dlgPresentedEnableUser := core.AutoState[bool](wnd)
	dlgPresentedDisableUser := core.AutoState[bool](wnd)
	dlgPresentedVerifyUser := core.AutoState[bool](wnd)
	dlgPresentedDeleteUser := core.AutoState[bool](wnd)
	dlgPresentedRemoveUserFromX := core.AutoState[bool](wnd)
	var dialogs []core.View
	dialogs = append(dialogs,

		ui.Lazy(func() core.View {
			if !dlgPresentedEnableUser.Get() {
				return nil
			}

			return alert.Dialog("Nutzer aktivieren", ui.Text("Sollen die ausgewählten Nutzer alle aktiviert werden?"), dlgPresentedEnableUser, alert.Cancel(nil), alert.Save(func() (close bool) {
				for _, state := range userStates {
					if state.selected.Get() {
						if err := useCases.MyCircleUserUpdateStatus(wnd.Subject(), circle.ID, state.usr.ID, user.Enabled{}); err != nil {
							alert.ShowBannerError(wnd, err)
							return false
						}
					}
				}

				return true
			}))
		}),

		ui.Lazy(func() core.View {
			if !dlgPresentedDisableUser.Get() {
				return nil
			}

			return alert.Dialog("Nutzer deaktivieren", ui.Text("Sollen die ausgewählten Nutzer alle deaktiviert werden?"), dlgPresentedDisableUser, alert.Cancel(nil), alert.Save(func() (close bool) {
				for _, state := range userStates {
					if state.selected.Get() {
						if err := useCases.MyCircleUserUpdateStatus(wnd.Subject(), circle.ID, state.usr.ID, user.Disabled{}); err != nil {
							alert.ShowBannerError(wnd, err)
							return false
						}
					}
				}

				return true
			}))
		}),

		ui.Lazy(func() core.View {
			if !dlgPresentedVerifyUser.Get() {
				return nil
			}

			return alert.Dialog("Nutzer verifizieren", ui.Text("Sollen die E-Mail Adressen der ausgewählten Nutzer ohne weitere Prüfung als verifiziert markiert werden?"), dlgPresentedVerifyUser, alert.Cancel(nil), alert.Save(func() (close bool) {
				for _, state := range userStates {
					if state.selected.Get() {
						if err := useCases.MyCircleUserVerified(wnd.Subject(), circle.ID, state.usr.ID, true); err != nil {
							alert.ShowBannerError(wnd, err)
							return false
						}
					}
				}

				return true
			}))
		}),

		ui.Lazy(func() core.View {
			if !dlgPresentedDeleteUser.Get() {
				return nil
			}

			return alert.Dialog("Nutzer löschen", ui.Text("Sollen die ausgewählten Nutzer unwiderruflich aus dem System entfernt werden?"), dlgPresentedDeleteUser, alert.Cancel(nil), alert.Save(func() (close bool) {
				for _, state := range userStates {
					if state.selected.Get() {
						if err := useCases.MyCircleUserRemove(wnd.Subject(), circle.ID, state.usr.ID); err != nil {
							alert.ShowBannerError(wnd, err)
							return false
						}
					}
				}

				return true
			}))
		}),

		ui.Lazy(func() core.View {
			if !dlgPresentedRemoveUserFromX.Get() {
				return nil
			}

			return alert.Dialog("Nutzerzuordnung entfernen", ui.Text("Sollen die ausgewählten Nutzer aus dieser Zuordnung entfernt werden? Der Nutzer verbleibt dabei im System."), dlgPresentedRemoveUserFromX, alert.Cancel(nil), alert.Delete(func() {
				var tmp []user.User
				for _, state := range userStates {
					if state.selected.Get() {
						tmp = append(tmp, state.usr)
					}
				}

				removeUser(tmp)
			}))
		}),
	)

	var groups []ui.TMenuGroup
	if hasUserMenu {
		var items []ui.TMenuItem
		if circle.CanDelete {
			items = append(items, ui.MenuItem(func() {
				dlgPresentedDeleteUser.Set(true)
			}, ui.Text("Löschen").FullWidth().TextAlignment(ui.TextAlignStart)))
		}

		if circle.CanEnable {
			items = append(items, ui.MenuItem(func() {
				dlgPresentedEnableUser.Set(true)
			}, ui.Text("Aktivieren").FullWidth().TextAlignment(ui.TextAlignStart)))
		}

		if circle.CanDisable {
			items = append(items, ui.MenuItem(func() {
				dlgPresentedDisableUser.Set(true)
			}, ui.Text("Deaktivieren").FullWidth().TextAlignment(ui.TextAlignStart)))
		}

		if circle.CanVerify {
			items = append(items, ui.MenuItem(func() {
				dlgPresentedVerifyUser.Set(true)
			}, ui.Text("E-Mail verifizieren").FullWidth().TextAlignment(ui.TextAlignStart)))
		}

		groups = append(groups, ui.MenuGroup(items...))
	}

	if hasUserMenu && dlgPresentedRemoveUserFromX != nil && removeUser != nil {
		groups = append(groups, ui.MenuGroup(
			ui.MenuItem(func() {
				dlgPresentedRemoveUserFromX.Set(true)
			}, ui.Text("Zuordnung entfernen").FullWidth().TextAlignment(ui.TextAlignStart)),
		))
	}

	dialogs = append(dialogs, ui.SecondaryButton(nil).Title(fmt.Sprintf("Aktion für %d Nutzer ...", selectedCount)))
	return ui.Menu(
		ui.HStack(
			dialogs...,
		),
		groups...,
	)
}
