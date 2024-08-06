package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui2"
	"go.wdy.de/nago/presentation/ui2/tracking"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.Component(".", func(wnd core.Window) core.View {
			return VStack(
				tracking.SupportRequestDialog(wnd),
				PrimaryButton(func() {
					wnd.ImportFiles(core.ImportFilesOptions{
						Multiple: true,
						OnCompletion: func(files []core.File) {
							for _, file := range files {
								fmt.Println(file.Name())
							}
						},
					})
				}).Title("Select Files"),
			).Frame(Frame{}.MatchScreen())
		})
	}).Run()
}
