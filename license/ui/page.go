package licenseui

import (
	"go.wdy.de/nago/license"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
)

func LicenseOverviewPage(wnd core.Window, statusCalculator license.CalculateStatus) core.View {
	status, err := statusCalculator(wnd.Subject())
	if err != nil {
		return alert.BannerError(err)
	}

	return ui.VStack(
		ui.WindowTitle("Lizenzübersicht"),
		ui.H2(""),
		staticLicenseTable(status),
	)
}

func staticLicenseTable(status license.Status) core.View {
	return nil
}
