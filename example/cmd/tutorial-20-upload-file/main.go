package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {
			return VStack(
				alert.BannerMessages(wnd),
				PrimaryButton(func() {
					wnd.ImportFiles(core.ImportFilesOptions{
						Multiple: true,
						OnCompletion: func(files []core.File) {
							for _, file := range files {
								fmt.Println(file.Name())
							}
							alert.ShowBannerMessage(wnd, alert.Message{
								Title:   "super",
								Message: "das hat vll. geklappt",
								Intent:  alert.IntentOk,
							})
						},
					})
				}).Title("Select Files"),
			).Frame(Frame{}.MatchScreen())
		})
	}).Run()
}
