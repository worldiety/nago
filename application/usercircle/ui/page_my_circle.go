package uiusercircles

import (
	"fmt"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	uiuser "go.wdy.de/nago/application/user/ui"
	"go.wdy.de/nago/application/usercircle"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/data/rquery"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/avatar"
	"go.wdy.de/nago/presentation/ui/list"
	"go.wdy.de/nago/presentation/ui/picker"
	"os"
	"slices"
)

type userState struct {
	usr      user.User
	selected *core.State[bool]
	visible  bool
}

func PageMyCircle(wnd core.Window, useCases usercircle.UseCases, findRoleById role.FindByID, findGroupById group.FindByID) core.View {
	id := usercircle.ID(wnd.Values()["id"])
	optCircle, err := useCases.FindByID(wnd.Subject(), id)
	if err != nil {
		return alert.BannerError(err)
	}

	if optCircle.IsNone() {
		return alert.BannerError(os.ErrNotExist)
	}

	circle := optCircle.Unwrap()

	allUsers, err := xslices.Collect2(useCases.MyCircleMembers(wnd.Subject(), circle.ID))
	if err != nil {
		return alert.BannerError(err)
	}

	if len(allUsers) == 0 {
		return alert.Banner("Keine Nutzer vorhanden", "In diesem Nutzerkreis sind derzeit keine Nutzer vorhanden.")
	}

	allUserStates := make([]userState, 0, len(allUsers))

	selectAllState := core.AutoState[bool](wnd).Observe(func(newValue bool) {
		for idx, state := range allUserStates {
			if state.visible {
				allUserStates[idx].selected.Set(newValue)
			}
		}
	})

	recalcSelectAllCheckbox := func() {
		allVisibleSelected := true
		for _, state := range allUserStates {
			if state.visible {
				if !state.selected.Get() {
					allVisibleSelected = false
					break
				}
			}
		}

		selectAllState.Set(allVisibleSelected)
	}

	for _, usr := range allUsers {
		allUserStates = append(allUserStates, userState{
			usr: usr,
			selected: core.StateOf[bool](wnd, "circle-select-"+string(usr.ID)).Observe(func(newValue bool) {
				recalcSelectAllCheckbox()
			}),
		})
	}

	searchState := core.AutoState[string](wnd)

	searchState.Observe(func(newValue string) {
		searchPredicate := rquery.SimplePredicate[user.User](newValue)
		for idx, usr := range allUsers {
			if searchState.Get() == "" {
				allUserStates[idx].visible = true
			} else {
				allUserStates[idx].visible = searchPredicate(usr)
			}

		}

		recalcSelectAllCheckbox()
	})

	searchPredicate := rquery.SimplePredicate[user.User](searchState.Get())
	for idx, usr := range allUsers {
		if searchState.Get() == "" {
			allUserStates[idx].visible = true
		} else {
			allUserStates[idx].visible = searchPredicate(usr)
		}

	}

	visibleCount := 0
	selectedCount := 0
	for _, state := range allUserStates {
		if state.visible {
			visibleCount++
		}

		if state.selected.Get() {
			selectedCount++
		}
	}

	dlgPresentedUserUsr := core.AutoState[user.User](wnd)
	dlgPresentedUserDetails := core.AutoState[bool](wnd)

	return ui.VStack(
		ui.H1(circle.Name),
		ui.HStack(
			ui.Lazy(func() core.View {
				if !dlgPresentedUserDetails.Get() {
					return nil
				}

				usr := dlgPresentedUserUsr.Get()

				var tmpRoles []role.Role
				for _, rid := range usr.Roles {
					r, _ := findRoleById(user.SU(), rid)
					if r.IsSome() {
						tmpRoles = append(tmpRoles, r.Unwrap())
					}
				}

				var tmpGroups []group.Group
				for _, rid := range usr.Groups {
					r, _ := findGroupById(user.SU(), rid)
					if r.IsSome() {
						tmpGroups = append(tmpGroups, r.Unwrap())
					}
				}

				return alert.Dialog("Über", uiuser.ViewProfile(wnd, tmpRoles, tmpGroups, usr.Email, usr.Contact), dlgPresentedUserDetails, alert.MinWidth(ui.L560), alert.Back(nil), alert.Closeable())
			}),
			ui.Lazy(func() core.View {
				if selectedCount > 0 {
					return makeMenu(wnd, circle, selectedCount, allUserStates, useCases)
				}

				return ui.SecondaryButton(func() {

				}).Enabled(false).Title("keine Nutzer ausgewählt")
			}),
			ui.Spacer(),
			ui.ImageIcon(heroSolid.MagnifyingGlass),
			ui.TextField("", searchState.Get()).Style(ui.TextFieldReduced).InputValue(searchState),
		).Alignment(ui.Trailing).FullWidth(),
		list.List(slices.Collect(func(yield func(view core.View) bool) {
			for _, state := range allUserStates {
				usr := state.usr
				if state.visible {
					entry := list.Entry().
						Headline(usr.String()).
						SupportingText(string(usr.Email)).
						Trailing(ui.SecondaryButton(func() {
							dlgPresentedUserUsr.Set(state.usr)
							dlgPresentedUserDetails.Set(true)
						}).PreIcon(heroSolid.Eye)).
						Leading(ui.HStack(
							ui.Checkbox(state.selected.Get()).
								InputChecked(state.selected),
							avatar.TextOrImage(usr.String(), usr.Contact.Avatar),
						).Gap(ui.L8))

					if !yield(entry) {
						return
					}

				}
			}

		})...).
			Caption(ui.HStack(ui.Checkbox(selectAllState.Get()).InputChecked(selectAllState), ui.Text("alle angezeigten Nutzer wählen"))).
			Footer(ui.VStack(
				ui.Text(fmt.Sprintf("%d/%d Nutzer in diesem Kreis sichtbar", visibleCount, len(allUsers))),
			)).
			Frame(ui.Frame{}.FullWidth()),
	).Alignment(ui.Leading).Gap(ui.L8).FullWidth()
}

func makeMenu(wnd core.Window, circle usercircle.Circle, selectedCount int, userStates []userState, useCases usercircle.UseCases) core.View {
	hasUserMenu := circle.CanEnable || circle.CanDisable || circle.CanDelete || circle.CanVerify

	myRoles, err := xslices.Collect2(useCases.MyRoles(wnd.Subject(), circle.ID))
	if err != nil {
		return alert.BannerError(err)
	}

	myGroups, err := xslices.Collect2(useCases.MyGroups(wnd.Subject(), circle.ID))
	if err != nil {
		return alert.BannerError(err)
	}

	dlgPresentedAddRoles := core.AutoState[bool](wnd)
	dlgPresentedRemoveRoles := core.AutoState[bool](wnd)
	dlgPresentedAddGroups := core.AutoState[bool](wnd)
	dlgPresentedRemoveGroups := core.AutoState[bool](wnd)
	dlgPresentedEnableUser := core.AutoState[bool](wnd)
	dlgPresentedDisableUser := core.AutoState[bool](wnd)
	dlgPresentedVerifyUser := core.AutoState[bool](wnd)
	dlgPresentedDeleteUser := core.AutoState[bool](wnd)
	var dialogs []core.View
	dialogs = append(dialogs,
		rolePicker(wnd, myRoles, "Rollen hinzufügen", dlgPresentedAddRoles, func(roles []role.Role) error {
			for _, state := range userStates {
				if state.selected.Get() {
					if err := useCases.MyCircleRolesAdd(wnd.Subject(), circle.ID, state.usr.ID, identifiers(roles)...); err != nil {
						return err
					}
				}
			}

			return nil
		}),
		rolePicker(wnd, myRoles, "Rollen entfernen", dlgPresentedRemoveRoles, func(roles []role.Role) error {
			for _, state := range userStates {
				if state.selected.Get() {
					if err := useCases.MyCircleRolesRemove(wnd.Subject(), circle.ID, state.usr.ID, identifiers(roles)...); err != nil {
						return err
					}
				}
			}

			return nil
		}),
		groupPicker(wnd, myGroups, "Gruppen hinzufügen", dlgPresentedAddGroups, func(groups []group.Group) error {
			for _, state := range userStates {
				if state.selected.Get() {
					if err := useCases.MyCircleGroupsAdd(wnd.Subject(), circle.ID, state.usr.ID, identifiers(groups)...); err != nil {
						return err
					}
				}
			}

			return nil
		}),

		groupPicker(wnd, myGroups, "Gruppen entfernen", dlgPresentedRemoveGroups, func(groups []group.Group) error {
			for _, state := range userStates {
				if state.selected.Get() {
					if err := useCases.MyCircleGroupsRemove(wnd.Subject(), circle.ID, state.usr.ID, identifiers(groups)...); err != nil {
						return err
					}
				}
			}

			return nil
		}),

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

	hasRoleMenu := len(circle.Roles) > 0

	if hasRoleMenu {
		var items []ui.TMenuItem
		items = append(items,
			ui.MenuItem(func() {
				dlgPresentedAddRoles.Set(true)
			}, ui.Text("Rollen hinzufügen").FullWidth().TextAlignment(ui.TextAlignStart)),

			ui.MenuItem(func() {
				dlgPresentedRemoveRoles.Set(true)
			}, ui.Text("Rollen entfernen").FullWidth().TextAlignment(ui.TextAlignStart)),
		)

		groups = append(groups, ui.MenuGroup(items...))
	}

	hasGroupMenu := len(circle.Groups) > 0
	if hasGroupMenu {
		var items []ui.TMenuItem
		items = append(items,
			ui.MenuItem(func() {
				dlgPresentedAddGroups.Set(true)
			}, ui.Text("zu Gruppen hinzufügen").FullWidth().TextAlignment(ui.TextAlignStart)),

			ui.MenuItem(func() {
				dlgPresentedRemoveGroups.Set(true)
			}, ui.Text("aus Gruppen entfernen").FullWidth().TextAlignment(ui.TextAlignStart)),
		)

		groups = append(groups, ui.MenuGroup(items...))
	}

	dialogs = append(dialogs, ui.SecondaryButton(nil).Title(fmt.Sprintf("Aktion für %d Nutzer ...", selectedCount)))
	return ui.Menu(
		ui.HStack(
			dialogs...,
		),
		groups...,
	)
}

func rolePicker(wnd core.Window, roles []role.Role, title string, presented *core.State[bool], onSelected func([]role.Role) error) core.View {
	if !presented.Get() {
		return nil
	}
	selectedState := core.StateOf[[]role.Role](wnd, title)
	return alert.Dialog(
		title,
		picker.Picker[role.Role](title, roles, selectedState).
			Title(title).
			MultiSelect(true).
			Frame(ui.Frame{}.FullWidth()),
		presented,
		alert.Cancel(nil),
		alert.Save(func() (close bool) {
			if err := onSelected(roles); err != nil {
				alert.ShowBannerError(wnd, err)
				return false
			}
			return true
		}),
	)

}

func groupPicker(wnd core.Window, groups []group.Group, title string, presented *core.State[bool], onSelected func([]group.Group) error) core.View {
	if !presented.Get() {
		return nil
	}
	selectedState := core.StateOf[[]group.Group](wnd, title)
	return alert.Dialog(
		title,
		picker.Picker[group.Group](title, groups, selectedState).
			Title(title).
			MultiSelect(true).
			Frame(ui.Frame{}.FullWidth()),
		presented,
		alert.Cancel(nil),
		alert.Save(func() (close bool) {
			if err := onSelected(groups); err != nil {
				alert.ShowBannerError(wnd, err)
				return false
			}
			return true
		}),
	)

}

func identifiers[E data.Aggregate[ID], ID data.IDType](values []E) []ID {
	tmp := make([]ID, 0, len(values))
	for _, value := range values {
		tmp = append(tmp, value.Identity())
	}
	return tmp
}
