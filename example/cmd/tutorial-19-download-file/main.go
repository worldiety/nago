package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/pkg/blob/mem"
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
					// CAUTION: this must always be within an action, do not put this into the render tree
					err := wnd.SendFiles(core.FilesIter(mem.From(mem.Entries{
						"test.txt": []byte("hello world"),
					})))

					// this is just for illustration
					err = fmt.Errorf("this is an unhandled infrastructure test error: %w", err)
					if err != nil {
						tracking.RequestSupport(wnd, fmt.Errorf("cannot send files by doing this use case: %w", err))
					}
				}).Title("Download"),
			).Frame(Frame{}.MatchScreen())
		})
	}).Run()
}
