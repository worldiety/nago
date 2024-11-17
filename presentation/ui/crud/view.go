package crud

import (
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"slices"
)

func View[Entity data.Aggregate[ID], ID data.IDType](opts TOptions[Entity, ID]) ui.DecoredView {
	var dataView core.View

	if opts.viewMode == ViewStyleDefault {
		dataView = ui.ViewThatMatches(opts.wnd,
			ui.SizeClass(core.SizeClassSmall, Cards[Entity, ID](opts).Frame(ui.Frame{MaxWidth: ui.L480}.FullWidth())),
			ui.SizeClass(core.SizeClassMedium, Table[Entity, ID](opts).Frame(ui.Frame{}.FullWidth())),
		)
	} else {
		dataView = List(opts)
	}

	return ui.VStack(
		ui.HStack(
			ui.VStack(
				ui.VStack(
					ui.Text(opts.title).Font(ui.Title),
					ui.HLineWithColor(ui.ColorAccent),
				).Padding(ui.Padding{Bottom: ui.L16}),
			),
			ui.Spacer(),
			ui.HStack(slices.Collect[core.View](func(yield func(core.View) bool) {
				yield(ui.TextField("", opts.queryState.String()).InputValue(opts.queryState).Style(ui.TextFieldReduced))
				for _, action := range opts.actions {
					yield(action)
				}
			})...,
			).Padding(ui.Padding{Bottom: ui.L16}),
		).FullWidth(),
		dataView,
	).Frame(ui.Frame{MinWidth: ui.L400})
}
