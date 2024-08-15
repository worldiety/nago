package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {
			date := core.AutoState[ora.Date](wnd)
			if date.Get().Zero() {
				date.Set(ora.Date{1, 6, 2024})
			}
			start := core.AutoState[ora.Date](wnd)
			if start.Get().Zero() {
				start.Set(ora.Date{2, 7, 2024})
			}
			end := core.AutoState[ora.Date](wnd)
			if end.Get().Zero() {
				end.Set(ora.Date{20, 7, 2024})
			}
			showAlert := core.AutoState[bool](wnd)

			return VStack(
				alert.Dialog("Achtung", fmt.Sprintf("Deine Eingabe: %v, start=%v end=%v", date, start, end), showAlert, alert.Ok()),
				SingleDatePicker("Geburtstag", date.Get(), date),

				RangeDatePicker("Urlaub", start.Get(), start, end.Get(), end),
				PrimaryButton(func() {
					showAlert.Set(true)
				}).Title("Check"),
			).Gap(L16).
				Frame(Frame{}.MatchScreen())
		})
	}).Run()
}
