package uisecret

import (
	"fmt"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/avatar"
	"go.wdy.de/nago/presentation/ui/cardlayout"
	"go.wdy.de/nago/presentation/ui/form"
	"go.wdy.de/nago/presentation/ui/picker"
	"reflect"
	"slices"
)

func EditSecretPage(
	wnd core.Window,
	pages Pages,
	deleteSecret secret.DeleteMySecretByID,
	findSecret secret.FindMySecretByID,
	updateSecret secret.UpdateMyCredentials,
	updateSecretGroups secret.UpdateMySecretGroups,
	findMyGroups group.FindMyGroups,
) core.View {
	id := secret.ID(wnd.Values()["id"])
	optScr, err := findSecret(wnd.Subject(), id)
	if err != nil {
		return alert.BannerError(err)
	}

	if optScr.IsNone() {
		return alert.Banner("Nicht gefunden", "Das Secret ist nicht mehr verfügbar.")
	}

	scr := optScr.Unwrap()

	initialCredentialValue := scr.Credentials
	state := core.AutoState[secret.Credentials](wnd).Init(func() secret.Credentials {
		return scr.Credentials
	})

	var availGroups []group.Group
	var initalSelected []group.Group
	for grp, err := range findMyGroups(wnd.Subject()) {
		if err != nil {
			return alert.BannerError(err)
		}

		availGroups = append(availGroups, grp)
		if slices.Contains(scr.Groups, grp.ID) {
			initalSelected = append(initalSelected, grp)
		}
	}

	selectedGroups := core.AutoState[[]group.Group](wnd).Init(func() []group.Group {
		return initalSelected
	})

	spec := newCredentialTypeSpec(reflect.TypeOf(scr.Credentials))
	logo := spec.LogoView()
	if avtLogo, ok := logo.(avatar.TAvatar); ok {
		logo = cardlayout.Card(spec.name).
			Style(cardlayout.TitleCompact).
			Body(
				ui.VStack(
					avtLogo.Size(ui.L80),
					ui.Text(spec.description).TextAlignment(ui.TextAlignCenter),
				).Gap(ui.L16).
					FullWidth(),
			).
			Frame(ui.Frame{}.FullWidth())
	}

	dlgPresented := core.AutoState[bool](wnd)
	return ui.VStack(
		alert.Dialog("Secret löschen?", ui.Text("Soll das Secret wirklich entfernt werden?"), dlgPresented, alert.Delete(func() {
			if err := deleteSecret(wnd.Subject(), id); err != nil {
				alert.ShowBannerError(wnd, err)
				return
			}

			wnd.Navigation().ResetTo(pages.Vault, nil)
		}), alert.Cancel(nil)),
		ui.H1("Secret bearbeiten"),

		ui.HStack(ui.TertiaryButton(func() {
			dlgPresented.Set(true)
		}).PreIcon(heroSolid.Trash).Title("Secret löschen")).
			FullWidth().
			Alignment(ui.Trailing),

		logo,

		cardlayout.Card("Credentials").
			Style(cardlayout.TitleCompact).
			Body(form.Auto[secret.Credentials](form.AutoOptions{}, state)).
			Frame(ui.Frame{}.FullWidth()),

		groupEditor(wnd, availGroups, selectedGroups),
		ui.HLineWithColor(ui.ColorAccent),
		ui.HStack(
			ui.SecondaryButton(func() {
				if initialCredentialValue.IsZero() {
					if err := deleteSecret(wnd.Subject(), id); err != nil {
						alert.ShowBannerError(wnd, err)
						return
					}
				}

				wnd.Navigation().ResetTo(pages.Vault, nil)
			}).Title("Abbrechen"),
			ui.PrimaryButton(func() {
				if err := updateSecret(wnd.Subject(), id, state.Get()); err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}

				var gids []group.ID
				for _, grp := range selectedGroups.Get() {
					gids = append(gids, grp.ID)
				}

				if err := updateSecretGroups(wnd.Subject(), id, gids); err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}

				fmt.Println(state.Get())
				wnd.Navigation().ForwardTo(pages.Vault, nil)
			}).Title("Speichern"),
		).
			FullWidth().
			Gap(ui.L8).
			Alignment(ui.Trailing),
	).Gap(ui.L16).Alignment(ui.Leading).Frame(ui.Frame{Width: ui.Full, MaxWidth: ui.L560})
}

func groupEditor(wnd core.Window, availGroups []group.Group, selectedGroups *core.State[[]group.Group]) core.View {
	return cardlayout.Card("Veröffentlichungen").
		Body(ui.VStack(
			picker.Picker[group.Group]("Gruppen", availGroups, selectedGroups).
				MultiSelect(true).
				Frame(ui.Frame{}.FullWidth()),
		).FullWidth()).
		Style(cardlayout.TitleCompact).
		Frame(ui.Frame{}.FullWidth())
}
