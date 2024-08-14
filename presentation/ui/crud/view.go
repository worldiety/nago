package crud

import (
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"slices"
)

func View[Entity data.Aggregate[ID], ID data.IDType](opts TOptions[Entity, ID]) ui.DecoredView {

	return ui.VStack(
		ui.Box(ui.BoxLayout{
			Trailing: ui.HStack(slices.Collect[core.View](func(yield func(core.View) bool) {
				yield(ui.TextField("", opts.queryState.String()).InputValue(opts.queryState).Style(ui.TextFieldReduced))
				for _, action := range opts.actions {
					yield(action)
				}
			})...).Gap(ui.L4),
			Leading: ui.Text(opts.title).Font(ui.Title),
		}).Frame(ui.Frame{Height: ui.L80}.FullWidth()),

		ui.ViewThatMatches(opts.wnd,
			ui.SizeClass(core.SizeClassSmall, Cards[Entity, ID](opts).Frame(ui.Frame{MaxWidth: ui.L480}.FullWidth())),
			ui.SizeClass(core.SizeClassMedium, Table[Entity, ID](opts).Frame(ui.Frame{}.FullWidth())),
		),
	).Frame(ui.Frame{MinWidth: ui.L400})
}
