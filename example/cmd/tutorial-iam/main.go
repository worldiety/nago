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
func sayHello(auth auth.Subject) (string, error) {
	if err := auth.Audit("de.worldiety.tutorial.say_hello"); err != nil {
		return "invalid", err
	}

	return "hello " + auth.Name(), nil
}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		iamCfg := application.IAMSettings{}
		iamCfg.Permissions.Permissions = iam.PermissionsFrom(Permissions())
		iamCfg = cfg.IAM(iamCfg)

		cfg.Component(".", func(wnd core.Window) core.View {
			return ui.NewVBox(func(vbox *ui.VBox) {
				vbox.Append(
					ui.NewActionButton("Berechtigungen", func() {
						wnd.Navigation().ForwardTo(iamCfg.Permissions.ID, nil)
					}),

					ui.NewActionButton("Benutzer", func() {
						wnd.Navigation().ForwardTo(iamCfg.Users.ID, nil)
					}),

					ui.NewActionButton("Anmelden", func() {
						wnd.Navigation().ForwardTo(iamCfg.Login.ID, nil)
					}),
					ui.NewActionButton("Abmelden", func() {
						wnd.Navigation().ForwardTo(iamCfg.Logout.ID, nil)
					}),
				)

				msg, err := sayHello(wnd.Subject())
				vbox.Append(ui.MakeText(fmt.Sprintf("%s:%v", msg, err)))
			})
		})

		cfg.OnDestroy(func() {
			fmt.Println("regular shutdown")
		})
	}).Run()
}
