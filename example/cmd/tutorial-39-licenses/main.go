package main

import (
	"fmt"
	"github.com/worldiety/enum"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/license"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"time"
)

var (
	_ = enum.Variant[secret.Credentials, secret.Jira]()
	_ = enum.Variant[secret.Credentials, secret.BookStack]()
)

var licensePucBasic = license.UserLicense{ID: "de.worldiety.puc.license.user.chat", Name: "PUC Basic License", MaxUsers: 10, Url: "https://www.worldiety.de/loesungen/puc"}
var licensePucImage = license.UserLicense{ID: "de.worldiety.puc.license.user.img", Name: "PUC Image License", MaxUsers: 5, Url: "https://www.worldiety.de/loesungen/puc"}
var licensePucJira = license.AppLicense{
	ID:          "de.worldiety.puc.license.app.jira",
	Name:        "PUC Jira License",
	Description: "Hiermit erhält PUC grundsätzlich Zugriff auf Jira Instanzen.",
	Url:         "https://www.worldiety.de/loesungen/puc#quellenangabe",
	Incentive:   "mailto:einkauf@worldiety.de?subject=PUC%20JIRA%20Lizenz&body=Liebes%20PUC%20Team%2C%0A%0Aich%20muss%20unbedingt%20die%20JIRA%20Integration%20freigeschaltet%20bekommen.%0A%0AViele%20Gr%C3%BC%C3%9Fe",
}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		// declare some hardcoded licenses and insert them at startup
		users := std.Must(cfg.UserManagement())
		licenses := std.Must(cfg.LicenseManagement())

		// note that user-license MaxUser number is not reset, if upserted
		std.Must(licenses.UseCases.PerUser.Upsert(users.UseCases.SysUser(), licensePucBasic))
		std.Must(licenses.UseCases.PerUser.Upsert(users.UseCases.SysUser(), licensePucImage))

		// note that app-license enabled flag is not reset, if upserted
		std.Must(licenses.UseCases.PerApp.Upsert(users.UseCases.SysUser(), licensePucJira))

		std.Must(cfg.BackupManagement())

		std.Must(cfg.MailManagement())
		std.Must(cfg.SessionManagement())

		std.Must(std.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "8fb8724f-e604-444c-9671-58d07dd76164"))

		cfg.SetDecorator(cfg.NewScaffold().
			Logo(heroSolid.AcademicCap).
			Decorator())

		cfg.RootView(".", cfg.DecorateRootView(func(wnd core.Window) core.View {

			return ui.VStack(

				ui.Text("User scoped enabled licenses:").Font(ui.Title),
				ui.VStack(
					ui.Each(wnd.Subject().Licenses(), func(t license.ID) core.View {
						return ui.Text(fmt.Sprintf("%v", t))
					})...,
				),
			)
		}))

	}).Run()
}
