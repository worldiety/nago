// main denotes an executable go package. If you don't know, what that means, go through the Go Tour first.
package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

// the main function of the program, which is like the java public static void main.
func main() {
	// we use the applications package to bootstrap our configuration
	application.Configure(func(cfg *application.Configurator) {
		// the application id is used to place files in the applications storage directory
		// e.g. in this case the directory is ~/.de.worldiety.tutorial
		// this is also used for logging and identification in complex service environments
		cfg.SetApplicationID("de.worldiety.tutorial")

		// we tell the application to deliver the build-in VueJS frontend. It will communicate over regular http REST
		// and websocket calls.
		cfg.Serve(vuejs.Dist())

		// Everything you see is based on scopes and components.
		// A scope is created by a client (e.g. the vuejs frontend above) and used as a scratch space for
		// allocating and freeing registered component factories.
		// Below you can see how to register a component factory, which just returns Text component to display the
		// obligatory 'hello world'. The dot . means the initial component, which represents the index page for the
		// web.
		cfg.Component(".", func(wnd core.Window) core.Component {
			// MakeText is a shortcut for a default text component
			return ui.MakeText("hello world")
		})
	}).
		// don't forget to call the run method, which starts the entire thing and blocks until finished
		Run()
}
