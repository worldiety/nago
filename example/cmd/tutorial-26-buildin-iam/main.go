// #[go.permission.generateTable]
package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

// sayHello greets everyone who has been authenticated.
// #[@Usecase]
// #[go.permission.audit]
func sayHello(auth auth.Subject) string {
	if err := auth.Audit("de.worldiety.tutorial.say_hello"); err != nil {
		return fmt.Sprintf("invalid: %v", err)
	}

	return "hello " + auth.Name()
}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		iamCfg := application.IAMSettings{}
		iamCfg.Permissions.Permissions = iam.PermissionsFrom(Permissions())
		iamCfg = cfg.IAM(iamCfg)

		cfg.RootView(".", func(wnd core.Window) core.View {
			return ui.VStack(
				ui.PrimaryButton(func() {
					wnd.Navigation().ForwardTo(iamCfg.Permissions.ID, nil)
				}).Title("Berechtigungen"),

				ui.PrimaryButton(func() {
					wnd.Navigation().ForwardTo(iamCfg.Users.ID, nil)
				}).Title("Benutzer"),

				ui.PrimaryButton(func() {
					wnd.Navigation().ForwardTo(iamCfg.Login.ID, nil)
				}).Title("Anmelden"),

				ui.PrimaryButton(func() {
					wnd.Navigation().ForwardTo(iamCfg.Logout.ID, nil)
				}).Title("Abmelden"),

				ui.Text(fmt.Sprintf("%s:%v", sayHello(wnd.Subject()))),
			).Gap(ui.L16).Frame(ui.Frame{}.MatchScreen())
		})

		cfg.OnDestroy(func() {
			fmt.Println("regular shutdown")
		})
	}).Run()
}
