package main

import (
	"context"
	_ "embed"
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"log/slog"
	"time"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.Component(".", func(wnd core.Window) core.View {
			seconds := core.AutoState[int](wnd)

			core.OnAppear(wnd, "my-time", func(ctx context.Context) {
				slog.Info("launched")
				for {
					if ctx.Err() != nil {
						slog.Error("my-timer has exited")
						break // exit
					}

					time.Sleep(time.Second)

					// states are thread safe for setting and getting
					seconds.Set(seconds.Get() + 1)

					slog.Info("my seconds", slog.Int("secs", seconds.Get()))

					// Note, that this endless-loop is one of the rare situations, where you have to
					// issue a manual invalidation, otherwise the window does not know when to render.
					// On the other hand, if this func is done, an invalidation is triggered automatically.
					wnd.Invalidate()
				}
				slog.Info("exit")
			})

			customState := fmt.Sprintf("state-%v", seconds)
			core.OnDisappear(wnd, customState, func(ctx context.Context) {
				slog.Info("disappeared", "old", customState, "active", seconds)
			})

			return ui.Text(fmt.Sprintf("seconds: %v", seconds))

		})
	}).Run()
}
