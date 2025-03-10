package uiuser

import (
	"go.wdy.de/nago/application/theme"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/cardlayout"
)

func PageSelfRegister(wnd core.Window) core.View {
	rows := 2
	if wnd.Info().SizeClass > core.SizeClassSmall {
		rows = 1
	}

	userSettings := core.GlobalSettings[user.Settings](wnd)
	_ = userSettings

	themeSettings := core.GlobalSettings[theme.Settings](wnd)

	return ui.VStack( //scaffold replacement
		ui.WindowTitle("Konto erstellen"),
		cardlayout.Card("").Body(
			ui.Grid(
				ui.GridCell(ui.VStack(
					ui.Image().
						Adaptive(themeSettings.AppIconLight, themeSettings.AppIconDark).
						Frame(ui.Frame{}.Size(ui.L48, ui.L48)),

					ui.Space(ui.L16),
					ui.Text(wnd.Application().Name()+"-Konto").Font(ui.Title),
					ui.Text("erstellen").Font(ui.Title),
				).Alignment(ui.TopLeading)),

				ui.GridCell(ui.VStack(
					ui.Space(ui.L48),
					ui.Space(ui.L16),
					ui.TextField("Vorname", "").FullWidth(),
					ui.TextField("Nachname", "").FullWidth(),
				).FullWidth().Gap(ui.L8)),
			).Gap(ui.L16).Rows(rows).FullWidth(),
		).Padding(ui.Padding{}.All(ui.L40)).
			Frame(ui.Frame{MaxWidth: ui.L880}.FullWidth()).
			Footer(ui.PrimaryButton(func() {

			}).Title("weiter")),
	).Frame(ui.Frame{}.MatchScreen())
}
