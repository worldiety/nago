// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiadmin

import (
	"go.wdy.de/nago/application/admin"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/application/slug"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/cardlayout"
)

type Pages struct {
	AdminCenter core.NavigationPath
}

func AdminCenter(wnd core.Window, queryGroups admin.QueryGroups) core.View {
	if !wnd.Subject().Valid() {
		return alert.BannerError(user.InvalidSubjectErr)
	}

	query := core.AutoState[string](wnd)

	adminGroups := queryGroups(wnd.Subject(), query.Get())

	isSmall := wnd.Info().SizeClass <= core.SizeClassSmall

	var viewBuilder xslices.Builder[core.View]
	viewBuilder.Append(
		ui.H1("Einstellungen"),

		ui.IfFunc(isSmall, func() core.View {
			return ui.VStack(
				ui.ImageIcon(heroSolid.MagnifyingGlass),
				ui.TextField("", query.Get()).
					InputValue(query).
					Style(ui.TextFieldReduced).
					FullWidth(),
				ui.Space(ui.L16),
			).Alignment(ui.Trailing).
				FullWidth()
		}),

		ui.IfFunc(!isSmall, func() core.View {
			return ui.HStack(
				ui.ImageIcon(heroSolid.MagnifyingGlass),
				ui.TextField("", query.Get()).
					InputValue(query).
					Style(ui.TextFieldReduced),
			).Alignment(ui.Trailing).
				FullWidth()
		}),
	)

	for _, grp := range adminGroups {
		viewBuilder.Append(ui.H2(grp.Title))
		var cardLayoutViews xslices.Builder[core.View]
		for _, entry := range grp.Entries {
			anchor := entry.ID
			if len(anchor) == 0 {
				anchor = slug.Slugify(grp.Title, entry.Title)
			}

			cardLayoutViews.Append(
				cardlayout.Card(entry.Title).
					ID(anchor).
					Body(ui.Text(entry.Text)).
					Footer(
						ui.SecondaryButton(func() {
							currentPath := wnd.Path()
							currentValues := wnd.Values()
							currentValues = currentValues.Put("#", anchor)

							wnd.Navigation().Replace(currentPath, currentValues)
							wnd.Navigation().ForwardTo(entry.Target, entry.TargetParams)
						}).Title(rstring.ActionSelect.Get(wnd)),
					),
			)
		}

		viewBuilder.Append(
			cardlayout.Layout(cardLayoutViews.Collect()...).Padding(ui.Padding{Bottom: ui.L80}),
		)

	}

	return ui.VStack(
		viewBuilder.Collect()...,
	).FullWidth().Alignment(ui.Leading)

}
