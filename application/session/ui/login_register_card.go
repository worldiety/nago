package uisession

import (
	"go.wdy.de/nago/application/image"
	httpimage "go.wdy.de/nago/application/image/http"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/cardlayout"
)

func LoginRegisterCard(wnd core.Window, logoLight, logoDark image.ID, title, subtitle string, contentRight core.View, actions ...core.View) core.View {
	logoImg := ui.IfFunc(logoLight != "" || logoDark != "", func() core.View {
		dark := httpimage.URI(logoDark, image.FitCover, 512, 512)
		light := httpimage.URI(logoLight, image.FitCover, 512, 512)
		return ui.VStack(
			ui.VStack(
				ui.Image().URIAdaptive(light, dark).ObjectFit(ui.FitContain).Frame(ui.Frame{Height: ui.L64}),
			).
				Action(func() {
					wnd.Navigation().ForwardTo(".", nil)
				}),
		).FullWidth()
	})

	content := ui.Grid(
		ui.GridCell(
			ui.VStack(
				logoImg,
				ui.VStack(
					ui.Text(title).Font(ui.HeadlineSmall).Hyphens(ui.HyphensAuto),
					ui.Text(subtitle),
				).Alignment(ui.Leading).FullWidth(),
			).Alignment(ui.TopLeading).Gap(ui.L32).FullWidth(),
		).Alignment(ui.Leading),
		ui.GridCell(contentRight),
	).
		Gap(ui.L32).
		Columns(1).
		FullWidth().
		Heights("auto")

	card := cardlayout.Card("").
		Body(
			ui.VStack(
				content,
			).Gap(ui.L16).FullWidth(),
		).
		Padding(ui.Padding{Top: ui.L48, Bottom: ui.L32}.Horizontal(ui.L48)).
		Frame(ui.Frame{MaxWidth: ui.L560}.FullWidth())

	hasActions := false
	for _, action := range actions {
		if action != nil {
			hasActions = true
			break
		}
	}
	if hasActions {
		card = card.Footer(ui.HStack(actions...).Gap(ui.L8))
	}

	return card
}
