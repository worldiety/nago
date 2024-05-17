package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/uix/xdialog"
	"go.wdy.de/nago/web/vuejs"
	"testing/fstest"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.Component(".", func(wnd core.Window) core.Component {
			return ui.NewButton(func(btn *ui.Button) {
				btn.Caption().Set("download")
				btn.Action().Set(func() {
					err := wnd.SendFiles(fstest.MapFS{
						"test.txt": &fstest.MapFile{
							Data: []byte("hello world"),
						},
					})
					xdialog.ErrorView("send files failed", err)
				})
			})
		})
	}).Run()
}
