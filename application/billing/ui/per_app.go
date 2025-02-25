package uibilling

import (
	"go.wdy.de/nago/application/billing"
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/presentation/core"
	heroOutline "go.wdy.de/nago/presentation/icons/hero/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/list"
	"go.wdy.de/nago/presentation/ui/tags"
	"slices"
)

func AppLicensePage(wnd core.Window, licenses billing.AppLicenses) core.View {
	appLicenses, err := xslices.Collect2(licenses(wnd.Subject()))
	if err != nil {
		return alert.BannerError(err)
	}

	return ui.VStack(
		ui.H1("Lizensierte Module"),
		list.List(
			ui.Each(slices.Values(appLicenses), func(t license.AppLicense) core.View {
				var tmp []core.View

				if t.Enabled {
					tmp = append(tmp, tags.ColoredTextPill(ui.ColorSemanticGood, "bereits gebucht"))
				} else {
					if t.Incentive != "" {
						tmp = append(tmp, ui.PrimaryButton(func() {
							wnd.Navigation().Open(core.URI(t.Incentive))
						}).Title("jetzt anfragen"))
					}

				}

				if t.Url != "" {
					tmp = append(tmp, ui.TertiaryButton(func() {
						core.HTTPOpen(wnd.Navigation(), core.URI(t.Url), "_blank")
					}).PreIcon(heroOutline.InformationCircle))
				}

				entry := list.Entry().
					Headline(t.Name).
					SupportingText(t.Description).
					Trailing(ui.HStack(tmp...).Gap(ui.L8))

				if t.Enabled {
					entry = entry.Leading(ui.ImageIcon(heroOutline.CheckCircle))
				} else {
					entry = entry.Leading(ui.ImageIcon(heroOutline.XMark))
				}

				return entry
			})...,
		).Frame(ui.Frame{}.FullWidth()),
	).Alignment(ui.Leading).FullWidth()
}
