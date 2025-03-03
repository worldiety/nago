package uiusercircles

import (
	"fmt"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/application/usercircle"
	"go.wdy.de/nago/pkg/data/rquery"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/avatar"
	"go.wdy.de/nago/presentation/ui/list"
	"os"
	"slices"
)

type userState struct {
	usr      user.User
	selected *core.State[bool]
	visible  bool
}

func PageMyCircle(wnd core.Window, useCases usercircle.UseCases) core.View {
	id := usercircle.ID(wnd.Values()["id"])
	optCircle, err := useCases.FindByID(wnd.Subject(), id)
	if err != nil {
		return alert.BannerError(err)
	}

	if optCircle.IsNone() {
		return alert.BannerError(os.ErrNotExist)
	}

	circle := optCircle.Unwrap()

	allUsers, err := xslices.Collect2(useCases.MyCircleMembers(wnd.Subject().ID(), circle.ID))
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

	return ui.VStack(
		ui.H1(circle.Name),
		ui.HStack(
			ui.Lazy(func() core.View {
				if selectedCount > 0 {
					return makeMenu(wnd, circle, selectedCount, allUserStates)
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

func makeMenu(wnd core.Window, circle usercircle.Circle, selectedCount int, userStates []userState) core.View {
	return ui.Menu(
		ui.SecondaryButton(nil).Title(fmt.Sprintf("Aktion für %d Nutzer ...", selectedCount)),
		ui.MenuGroup(
			ui.MenuItem(func() {
				fmt.Println("löschen")
			}, ui.Text("Nutzer Löschen").FullWidth().TextAlignment(ui.TextAlignStart)),
			ui.MenuItem(func() {
				fmt.Println("deaktivieren")
			}, ui.Text("Nutzer deaktivieren").FullWidth().TextAlignment(ui.TextAlignStart)),
		),

		ui.MenuGroup(
			ui.MenuItem(func() {
				fmt.Println("löschen2")
			}, ui.Text("Nutzer Löschen").FullWidth().TextAlignment(ui.TextAlignStart)),
			ui.MenuItem(func() {
				fmt.Println("deaktivieren2")
			}, ui.Text("Nutzer deaktivieren").FullWidth().TextAlignment(ui.TextAlignStart)),
		),
	)
}
