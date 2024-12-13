package uiadmin

import (
	"go.wdy.de/nago/application/admin"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/presentation/core"
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

	var viewBuilder xslices.Builder[core.View]
	viewBuilder.Append(
		ui.H1("Admin Center"),

		ui.HStack(
			ui.TextField("", query.Get()).
				InputValue(query).
				Style(ui.TextFieldReduced),
		).Alignment(ui.Trailing).
			FullWidth(),
	)

	for _, grp := range adminGroups {
		viewBuilder.Append(ui.H2(grp.Title))
		var cardLayoutViews xslices.Builder[core.View]
		for i, entry := range grp.Entries {
			cardLayoutViews.Append(
				cardlayout.Card(entry.Title).
					Body(ui.Text(entry.Text)).
					Footer(
						ui.IfElse(i == 0,
							ui.PrimaryButton(func() {
								wnd.Navigation().ForwardTo(entry.Target, nil)
							}).Title("Auswählen"),
							ui.SecondaryButton(func() {
								wnd.Navigation().ForwardTo(entry.Target, nil)
							}).Title("Auswählen"),
						),
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
