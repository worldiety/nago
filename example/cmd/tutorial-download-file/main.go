package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/pkg/blob/mem"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"log/slog"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.Component(".", func(wnd core.Window) core.View {
			return ui.NewButton(func(btn *ui.Button) {
				btn.Caption().Set("download")
				btn.Action().Set(func() {
					err := wnd.SendFiles(core.FilesIter(mem.From(mem.Entries{
						"test.txt": []byte("hello world"),
					})))
					if err != nil {
						slog.Error("cannot send files", slog.Any("err", err))
					}
				})
			})
		})
	}).Run()
}
