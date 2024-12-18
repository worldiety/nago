package tags

import "go.wdy.de/nago/presentation/ui"

func ColoredTextPill(color ui.Color, text string) ui.DecoredView {
	return ui.HStack(
		ui.Text(text).Color(ui.ColorBlack),
	).BackgroundColor(color).
		Padding(ui.Padding{}.Horizontal(ui.L8).Vertical(ui.L4)).
		Border(ui.Border{}.Radius(ui.L16))
}
