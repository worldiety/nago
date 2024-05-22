// #[go.permission.generateTable]
package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/auth/iam"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

type Authenticated interface {
	Audit(permission string) error
}

// sayHello greets everyone who has been authenticated.
// #[@Usecase]
// #[go.permission.audit]
func sayHello(auth Authenticated) (string, error) {
	if err := auth.Audit("de.worldiety.tutorial.say_hello"); err != nil {
		return "invalid", err
	}

	return "hello", nil
}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.IAM(application.IAMSettings{
			Permissions: application.Permissions{
				Permissions: iam.PermissionsFrom(Permissions()),
			},
		})

		cfg.Component(".", func(wnd core.Window) core.Component {
			return ui.NewVBox(func(vbox *ui.VBox) {
				vbox.Append(
					ui.NewActionButton("Berechtigungen", func() {
						wnd.Navigation().ForwardTo("iam/permissions", nil)
					}),

					ui.NewActionButton("Benutzer", func() {
						wnd.Navigation().ForwardTo("iam/users", nil)
					}),

					ui.NewActionButton("Anmelden", func() {
						wnd.Navigation().ForwardTo("iam/login", nil)
					}),
					ui.NewActionButton("Abmelden", func() {
						wnd.Navigation().ForwardTo("iam/logout", nil)
					}),
				)

				msg, err := sayHello(wnd.Subject())
				vbox.Append(ui.MakeText(fmt.Sprintf("%s:%v", msg, err)))
			})
		})
	}).Run()
}
