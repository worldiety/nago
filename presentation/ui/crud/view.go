// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package crud

import (
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"slices"
)

func View[Entity data.Aggregate[ID], ID data.IDType](opts TOptions[Entity, ID]) ui.DecoredView {
	var dataView core.View

	if opts.viewMode == ViewStyleDefault {
		dataView = ui.ViewThatMatches(opts.wnd,
			ui.SizeClass(core.SizeClassSmall, func() core.View { return Cards[Entity, ID](opts).Frame(ui.Frame{MaxWidth: ui.L480}.FullWidth()) }),
			ui.SizeClass(core.SizeClassMedium, func() core.View { return Table[Entity, ID](opts).Frame(ui.Frame{}.FullWidth()) }),
		)
	} else {
		dataView = List(opts)
	}

	searchbarAndActions := slices.Collect[core.View](func(yield func(core.View) bool) {
		yield(ui.ImageIcon(heroSolid.MagnifyingGlass))
		yield(ui.TextField("", opts.queryState.String()).InputValue(opts.queryState).Style(ui.TextFieldReduced))
		if len(opts.actions) > 0 {
			yield(ui.FixedSpacer(ui.L16, ""))
		}

		for _, action := range opts.actions {
			yield(action)
		}
	})

	isSmall := opts.wnd.Info().SizeClass <= core.SizeClassSmall

	return ui.VStack(
		ui.IfFunc(isSmall, func() core.View {
			return ui.VStack(
				ui.HStack(ui.If(opts.title != "", ui.H1(opts.title))).FullWidth().Alignment(ui.Leading),
				ui.HStack(searchbarAndActions...).Padding(ui.Padding{Bottom: ui.L16}),
			).FullWidth().Alignment(ui.Trailing)
		}),
		ui.IfFunc(!isSmall, func() core.View {
			return ui.HStack(
				ui.If(opts.title != "", ui.H1(opts.title)),
				ui.Spacer(),
				ui.HStack(searchbarAndActions...).Padding(ui.Padding{Bottom: ui.L16}),
			).FullWidth()
		}),

		dataView,
	).Frame(ui.Frame{MinWidth: ui.L400})
}
