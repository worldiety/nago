package uiuser

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/cardlayout"
)

func PageSelfRegister(wnd core.Window) core.View {
	return ui.VStack( //scaffold replacement

		cardlayout.Card("").Body(
			ui.VStack(
				ui.WindowTitle("Konto erstellen"),
				ui.Text("Konto erstellen"),
			),
		).Padding(ui.Padding{}.All(ui.L12)).
			Footer(ui.PrimaryButton(func() {

			}).Title("weiter")),
	).Frame(ui.Frame{}.MatchScreen())
}
