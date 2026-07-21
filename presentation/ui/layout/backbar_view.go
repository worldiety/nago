package layout

import (
	"github.com/worldiety/i18n"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"golang.org/x/text/language"
)

var (
	StrBack = i18n.MustString("backbar-view.back", i18n.Values{language.German: "Zurück", language.English: "Back"})
)

func WithBackButton(wnd core.Window, view core.View) core.View {
	return ui.VStack(
		ui.HStack(
			ui.TertiaryButton(func() {
				wnd.Navigation().Back()
			}).Title(StrBack.Get(wnd)).PreIcon(icons.ArrowLeft),
		).FullWidth().Alignment(ui.Leading).Padding(ui.Padding{}.Vertical(ui.L16)),
		view,
	).Gap(ui.L8).FullWidth()
}
