// #[go.permission.generateTable]
package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"time"
)

var myPermission = permission.Declare[SayHello]("de.worldiety.tutorial.say_hello", "Jeden Grüßen", "Diese Erlaubnis muss dem Nutzer zugewiesen werden.")

// SayHello greets everyone who has been authenticated.
type SayHello func(auth auth.Subject) string

func NewSayHello() SayHello {
	return func(auth auth.Subject) string {
		if err := auth.Audit(myPermission); err != nil {
			return fmt.Sprintf("invalid: %v", err)
		}

		return "hello " + auth.Name()
	}
}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		std.Must(cfg.Authentication())
		cfg.SetDecorator(cfg.NewScaffold().Decorator())

		std.Must(std.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "8fb8724f-e604-444c-9671-58d07dd76164"))

		sayHello := NewSayHello()

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			return ui.VStack(
				ui.Text(fmt.Sprintf("%s", sayHello(wnd.Subject()))),
			).Gap(ui.L16).Frame(ui.Frame{}.MatchScreen())
		})

	}).Run()
}
