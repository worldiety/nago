package crud

import (
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

func NewView[Entity data.Aggregate[ID], ID data.IDType](opts TOptions[Entity, ID]) ui.DecoredView {

	quickSearch := core.AutoState[string](opts.wnd).From(func() string {
		return opts.query
	})

	opts.queryState = quickSearch

	return ui.VStack(
		ui.Box(ui.BoxLayout{
			Trailing: ui.TextField("", quickSearch.String()).InputValue(quickSearch).Style(ui.TextFieldReduced),
			Leading:  ui.Text(opts.title).Font(ui.Title),
		}).Frame(ui.Frame{Height: ui.L80}.FullWidth()),

		ui.ViewThatMatches(opts.wnd,
			ui.SizeClass(core.SizeClassSmall, Cards[Entity, ID](opts).Frame(ui.Frame{MaxWidth: ui.L480}.FullWidth())),
			ui.SizeClass(core.SizeClassMedium, Table[Entity, ID](opts).Frame(ui.Frame{}.FullWidth())),
		),
	).Frame(ui.Frame{MinWidth: ui.L400})
}
