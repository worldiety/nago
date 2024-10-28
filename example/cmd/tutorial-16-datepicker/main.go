package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/pkg/xtime"
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
			date := core.AutoState[xtime.Date](wnd).Init(func() xtime.Date {
				return xtime.Date{Day: 1, Month: 6, Year: 2024}
			})

			start := core.AutoState[xtime.Date](wnd).Init(func() xtime.Date {
				return xtime.Date{Day: 2, Month: 7, Year: 2024}
			})

			end := core.AutoState[xtime.Date](wnd).Init(func() xtime.Date {
				return xtime.Date{Day: 20, Month: 7, Year: 2024}
			})

			showAlert := core.AutoState[bool](wnd)

			return VStack(
				alert.Dialog("Achtung", Text(fmt.Sprintf("Deine Eingabe: %v, start=%v end=%v", date, start, end)), showAlert, alert.Ok()),
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
