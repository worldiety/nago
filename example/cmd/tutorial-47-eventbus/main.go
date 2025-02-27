package main

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"log/slog"
)

type MyCustomEvent struct {
	Msg string
}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		option.MustZero(cfg.StandardSystems())

		cfg.SetDecorator(cfg.NewScaffold().
			Logo(ui.Image().Embed(heroSolid.AcademicCap).Frame(ui.Frame{}.Size(ui.L96, ui.L96))).
			Decorator())

		evtBus := cfg.EventBus()

		cfg.RootView(".", cfg.DecorateRootView(func(wnd core.Window) core.View {

			msgState := core.AutoState[string](wnd)
			wnd.AddDestroyObserver(events.SubscribeFor[MyCustomEvent](evtBus, func(evt MyCustomEvent) {
				msgState.Set("custom event received: " + evt.Msg)
			}))

			// open another tab and create a new user and watch how this message appears in the other tab.
			// Note, that you must "notify" the user, because otherwise the event will not occur (today).
			wnd.AddDestroyObserver(events.SubscribeFor[user.Created](evtBus, func(evt user.Created) {
				slog.Info("user created", "mail", evt.Email)
				msgState.Set("user was created: " + string(evt.Email))
			}))

			return ui.VStack(
				ui.Text("Last message from bus: "+msgState.Get()),
				ui.PrimaryButton(func() {
					evtBus.Publish(MyCustomEvent{Msg: "hello"})
				}).Title("send message"),
			).Frame(ui.Frame{}.MatchScreen())
		}))

	}).Run()
}
