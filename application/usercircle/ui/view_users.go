package uiusercircles

import (
	"fmt"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/application/user"
	uiuser "go.wdy.de/nago/application/user/ui"
	"go.wdy.de/nago/application/usercircle"
	"go.wdy.de/nago/pkg/data/rquery"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/pkg/xstrings"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/avatar"
	"go.wdy.de/nago/presentation/ui/list"
	"go.wdy.de/nago/presentation/ui/picker"
	"slices"
)

func viewUsers(wnd core.Window, subtitle string, useCases usercircle.UseCases, usrVisible func(usr user.User) bool, addUser func(users []user.User), removeUser func(users []user.User)) core.View {
	// security note: this gives us a lot protection in the UI, because if the subject is not a circle admin anymore, we will exit immediately
	circle, err := loadMyCircle(wnd, useCases)
	if err != nil {
		return alert.BannerError(err)
	}

	allCircleUsers, err := xslices.Collect2(useCases.MyCircleMembers(wnd.Subject(), circle.ID))
	if err != nil {
		return alert.BannerError(err)
	}

	if len(allCircleUsers) == 0 {
		return alert.Banner("Keine Nutzer vorhanden", "In diesem Nutzerkreis sind derzeit keine Nutzer vorhanden.")
	}

	var allUsers []user.User
	var availableUsers []user.User
	for _, circleUser := range allCircleUsers {
		if usrVisible(circleUser) {
			allUsers = append(allUsers, circleUser)
		} else {
			availableUsers = append(availableUsers, circleUser)
		}
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

	dlgPresentedAddUser := core.AutoState[bool](wnd)

	findRoleByID, _ := core.SystemService[role.FindByID](wnd.Application())
	findGroupByID, _ := core.SystemService[group.FindByID](wnd.Application())
	findLicenseByID, _ := core.SystemService[license.FindUserLicenseByID](wnd.Application())

	return ui.VStack(
		ui.H1(circle.Name+" / "+subtitle),
		ui.HStack(
			ui.Lazy(func() core.View {
				if !dlgPresentedAddUser.Get() {
					return nil
				}

				if len(availableUsers) == 0 {
					return alert.Dialog("Keine Nutzer verfügbar", ui.Text("Es sind bereits alle Nutzer des Kreises zugewiesen."), dlgPresentedAddUser, alert.Ok())
				}

				usersToAdd := core.AutoState[[]user.User](wnd)
				usersToAdd.Observe(addUser)
				return picker.Picker[user.User]("Verfügbare Nutzer", availableUsers, usersToAdd).
					WithDialogPresented(dlgPresentedAddUser).
					MultiSelect(true).
					Dialog()
			}),

			ui.Lazy(func() core.View {
				if !dlgPresentedUserDetails.Get() {
					return nil
				}

				usr := dlgPresentedUserUsr.Get()

				var tmpRoles []role.Role
				if findRoleByID != nil {
					for _, rid := range usr.Roles {
						r, _ := findRoleByID(user.SU(), rid)
						if r.IsSome() {
							tmpRoles = append(tmpRoles, r.Unwrap())
						}
					}
				}

				var tmpGroups []group.Group
				if findGroupByID != nil {
					for _, rid := range usr.Groups {
						r, _ := findGroupByID(user.SU(), rid)
						if r.IsSome() {
							tmpGroups = append(tmpGroups, r.Unwrap())
						}
					}
				}

				var tmpLicenses []license.UserLicense
				if findLicenseByID != nil {
					for _, rid := range usr.Licenses {
						r, _ := findLicenseByID(user.SU(), rid)
						if r.IsSome() {
							tmpLicenses = append(tmpLicenses, r.Unwrap())
						}
					}
				}

				return alert.Dialog("Über", uiuser.ViewProfile(wnd, tmpRoles, tmpGroups, tmpLicenses, usr.Email, usr.Contact), dlgPresentedUserDetails, alert.MinWidth(ui.L560), alert.Back(nil), alert.Closeable())
			}),
			ui.Lazy(func() core.View {
				if selectedCount > 0 {
					return makeMenu(wnd, circle, selectedCount, allUserStates, useCases, removeUser)
				}

				return ui.SecondaryButton(nil).Enabled(false).Title("keine Nutzer ausgewählt")
			}),
			ui.Spacer(),
			ui.ImageIcon(heroSolid.MagnifyingGlass),
			ui.TextField("", searchState.Get()).Style(ui.TextFieldReduced).InputValue(searchState),
			ui.If(addUser != nil, ui.PrimaryButton(func() {
				dlgPresentedAddUser.Set(true)
			}).Title("Nutzer zuordnen").AccessibilityLabel("Nutzer auswählen, um sie "+subtitle+" zuzuordnen."),
			),
		).Alignment(ui.Trailing).FullWidth(),
		list.List(slices.Collect(func(yield func(view core.View) bool) {
			for _, state := range allUserStates {
				usr := state.usr
				if state.visible {
					entry := list.Entry().
						Headline(xstrings.Join2(" ", usr.Contact.Firstname, usr.Contact.Lastname)).
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
				ui.Text(fmt.Sprintf("%d/%d Nutzer in dieser Zuordnung sichtbar", visibleCount, len(allUsers))),
			)).
			Frame(ui.Frame{}.FullWidth()),
	).Alignment(ui.Leading).Gap(ui.L8).FullWidth()
}
