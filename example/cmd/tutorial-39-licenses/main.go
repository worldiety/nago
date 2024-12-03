package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

var licensePucBasic = license.UserLicense{ID: "de.worldiety.puc.license.user.chat", Name: "PUC Basic License", MaxUsers: 10, Url: "https://www.worldiety.de/loesungen/puc"}
var licensePucImage = license.UserLicense{ID: "de.worldiety.puc.license.user.img", Name: "PUC Image License", MaxUsers: 5, Url: "https://www.worldiety.de/loesungen/puc"}
var licensePucJira = license.AppLicense{ID: "de.worldiety.puc.license.app.jira", Name: "PUC Jira License", Description: "Hiermit erhält PUC grundsätzlich Zugriff auf Jira Instanzen.", Url: "https://www.worldiety.de/loesungen/puc#quellenangabe"}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		// declare our licenses globally, this is also evaluated by IAM by default
		license.Global.StoreAll(license.FromEnv(
			licensePucBasic,
			licensePucImage,
			licensePucJira,
		)...)

		std.Must(cfg.MailManagement())
		
		cfg.SetDecorator(cfg.NewScaffold().Decorator())

		iamCfg := application.IAMSettings{}
		iamCfg.Decorator = cfg.Decorator()
		iamCfg = cfg.IAM(iamCfg)

		cfg.RootView(".", iamCfg.DecorateRootView(func(wnd core.Window) core.View {

			return ui.VStack(
				ui.Text("Global declared licenses:").Font(ui.Title),
				ui.VStack(
					ui.Each(license.Global.Values(), func(t license.License) core.View {
						return ui.Text(fmt.Sprintf("%v: %v", t.LicenseName(), t.Enabled()))
					})...,
				),
				ui.Text("User scoped enabled licenses:").Font(ui.Title),
				ui.VStack(
					ui.Each(wnd.Subject().Licenses(), func(t license.ID) core.View {
						lic, ok := license.Global.Load(t)
						if !ok {
							return ui.Text(fmt.Sprintf("missing id: %s", t))
						}

						return ui.Text(fmt.Sprintf("%v: %v", lic.LicenseName(), lic.Enabled()))
					})...,
				),
			)
		}))

	}).Run()
}
