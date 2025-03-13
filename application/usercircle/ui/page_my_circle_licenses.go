package uiusercircles

import (
	"fmt"
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/application/usercircle"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/list"
)

func PageMyCircleLicenses(wnd core.Window, pages Pages, useCases usercircle.UseCases, findLicByID license.FindUserLicenseByID) core.View {
	circle, err := loadMyCircle(wnd, useCases)
	if err != nil {
		return alert.BannerError(err)
	}

	var licenses []license.UserLicense
	for _, id := range circle.Licenses {
		optRole, err := findLicByID(user.SU(), id) // security note: we are allowed by user circle definition
		if err != nil {
			return alert.BannerError(err)
		}

		if optRole.IsNone() {
			continue
		}

		licenses = append(licenses, optRole.Unwrap())
	}

	return ui.VStack(
		ui.H1(circle.Name+" / Lizenzen"),
		list.List(ui.ForEach(licenses, func(t license.UserLicense) core.View {
			return list.Entry().
				Headline(t.Name).
				SupportingText(t.Description + fmt.Sprintf(" (max. %d Nutzer)", t.MaxUsers)).
				Trailing(ui.ImageIcon(heroSolid.ChevronRight))
		})...).OnEntryClicked(func(idx int) {
			lic := licenses[idx]
			wnd.Navigation().ForwardTo(pages.MyCircleLicensesUsers, core.Values{"circle": string(circle.ID), "license": string(lic.ID)})
		}).
			Caption(ui.Text("In diesem Kreis sichtbare Lizenzen")).
			Footer(ui.Text(fmt.Sprintf("%d Rollen sind zur Verwaltung verfügbar", len(licenses)))).
			Frame(ui.Frame{}.FullWidth()),
	).Alignment(ui.Leading).FullWidth()
}
