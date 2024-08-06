// #[go.permission.generateTable]
package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/uilegacy"
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
			return uilegacy.NewVBox(func(vbox *uilegacy.VBox) {
				vbox.Append(
					uilegacy.NewActionButton("Berechtigungen", func() {
						wnd.Navigation().ForwardTo(iamCfg.Permissions.ID, nil)
					}),

					uilegacy.NewActionButton("Benutzer", func() {
						wnd.Navigation().ForwardTo(iamCfg.Users.ID, nil)
					}),

					uilegacy.NewActionButton("Anmelden", func() {
						wnd.Navigation().ForwardTo(iamCfg.Login.ID, nil)
					}),
					uilegacy.NewActionButton("Abmelden", func() {
						wnd.Navigation().ForwardTo(iamCfg.Logout.ID, nil)
					}),
				)

				msg, err := sayHello(wnd.Subject())
				vbox.Append(uilegacy.MakeText(fmt.Sprintf("%s:%v", msg, err)))
			})
		})

		cfg.OnDestroy(func() {
			fmt.Println("regular shutdown")
		})
	}).Run()
}
