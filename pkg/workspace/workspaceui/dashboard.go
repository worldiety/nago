package workspaceui

import (
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/pkg/workspace"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/cardlayout"
	"slices"
)

type DashboardType struct {
	Icon        core.SVG
	Type        workspace.Type
	Name        string
	Description string
}

type DashboardOptions struct {
	Title string
	Types []DashboardType
	// OverviewListPath is forwarded with attached "type" parameter.
	OverviewListPath core.NavigationPath
}

func Dashboard(wnd core.Window, opts DashboardOptions) core.View {
	if !wnd.Subject().Valid() {
		return alert.BannerError(iam.InvalidSubjectError("not logged in"))
	}

	return ui.VStack(
		ui.H1(opts.Title),
		cardlayout.Layout(
			ui.Each(slices.Values(opts.Types), func(t DashboardType) core.View {
				return cardlayout.Card(t.Name).
					Body(ui.Text(t.Description)).
					Footer(ui.SecondaryButton(func() {
						wnd.Navigation().ForwardTo(opts.OverviewListPath, core.Values{"type": string(t.Type)})
					}).Title("Auswählen"))
			})...,
		),
	).Alignment(ui.Leading).FullWidth()

}
