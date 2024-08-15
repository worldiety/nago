// main denotes an executable go package. If you don't know, what that means, go through the Go Tour first.
package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

// the main function of the program, which is like the java public static void main.
func main() {
	// we use the applications package to bootstrap our configuration
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		var iamCfg application.IAMSettings
		iamCfg = cfg.IAM(iamCfg)

		cfg.RootView("blub", func(wnd core.Window) core.View {
			return PrimaryButton(func() {
				wnd.Navigation().ForwardTo(".", nil)
			}).Title("bla")
		})

		cfg.RootView(".", func(wnd core.Window) core.View {
			return Scaffold(ScaffoldAlignmentLeading).Body(VStack(

				PrimaryButton(func() {
					fmt.Println("btn1")
					wnd.Navigation().ForwardTo(".", nil)
				}).Title("btn1"),
				PrimaryButton(func() {
					fmt.Println("btn2")
					wnd.Navigation().ForwardTo(".", nil)
				}).Title("btn2"),
			).
				Frame(Frame{}.MatchScreen())).Menu(

				ParentScaffoldMenuEntry(wnd, heroSolid.AcademicCap, "bla",
					ForwardScaffoldMenuEntry(wnd, heroSolid.Plus, "blub1", "blub"),
					ForwardScaffoldMenuEntry(wnd, heroSolid.Plus, "blub2", "blub"),
					ForwardScaffoldMenuEntry(wnd, heroSolid.Plus, "blub3", "blub"),
				),
			)
		})
	}).
		// don't forget to call the run method, which starts the entire thing and blocks until finished
		Run()
}
