// main denotes an executable go package. If you don't know, what that means, go through the Go Tour first.
package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/ui2"
	"go.wdy.de/nago/web/vuejs"
)

// the main function of the program, which is like the java public static void main.
func main() {
	// we use the applications package to bootstrap our configuration
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.Component(".", func(wnd core.Window) core.View {
			return ui.VStack(func(vstack *ui.TVStack) {
				vstack.Frame(ora.Frame{Height: ora.ViewportHeight, Width: ora.Full})

				vstack.Of(
					ui.TextFrom("hello world"),
				)
			})

		})
	}).
		// don't forget to call the run method, which starts the entire thing and blocks until finished
		Run()
}
