package main

import (
	"fmt"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/form"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_99")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {

			type TfObj struct {
				Zeitraum xtime.TimeFrame
			}

			state := core.AutoState[TfObj](wnd).Init(func() TfObj {
				return TfObj{Zeitraum: xtime.TimeFrame{
					StartTime: xtime.Now(),
					EndTime:   xtime.Now(),
					Timezone:  "Local",
				}}
			})
			obj := state.Get()

			dStart := obj.Zeitraum.StartTime.Time(obj.Zeitraum.Timezone.Location()).Day()
			mStart := obj.Zeitraum.StartTime.Time(obj.Zeitraum.Timezone.Location()).Month()
			yStart := obj.Zeitraum.StartTime.Time(obj.Zeitraum.Timezone.Location()).Year()
			dateStart := fmt.Sprintf("%02d.%02d.%v", dStart, mStart, yStart)
			s := xtime.Date{
				Day:   dStart,
				Month: mStart,
				Year:  yStart,
			}

			dEnd := obj.Zeitraum.EndTime.Time(obj.Zeitraum.Timezone.Location()).Day()
			mEnd := obj.Zeitraum.EndTime.Time(obj.Zeitraum.Timezone.Location()).Month()
			yEnd := obj.Zeitraum.EndTime.Time(obj.Zeitraum.Timezone.Location()).Year()
			dateEnd := fmt.Sprintf("%02d.%02d.%v", dEnd, mEnd, yEnd)

			e := xtime.Date{
				Day:   dEnd,
				Month: mEnd,
				Year:  yEnd,
			}

			startHour := obj.Zeitraum.StartTime.Time(obj.Zeitraum.Timezone.Location()).Hour()
			startMinute := obj.Zeitraum.StartTime.Time(obj.Zeitraum.Timezone.Location()).Minute()
			endHour := obj.Zeitraum.EndTime.Time(obj.Zeitraum.Timezone.Location()).Hour()
			endMinute := obj.Zeitraum.EndTime.Time(obj.Zeitraum.Timezone.Location()).Minute()
			t := fmt.Sprintf("%02d:%02d – %02d:%02d", startHour, startMinute, endHour, endMinute)

			date := dateStart
			if e.After(s) {
				date += fmt.Sprintf(", %02d:%02d Uhr", startHour, startMinute)
				date += " bis "
				date += dateEnd
				date += fmt.Sprintf(", %02d:%02d Uhr", endHour, endMinute)
			} else {
				date += ", " + t + " Uhr"
			}

			return ui.VStack(
				ui.Space(ui.L40),
				ui.VStack(
					ui.Text("Zeitraum wählen").Font(ui.Font{Size: ui.L20, Weight: ui.DisplayAndLabelFontWeight}),
					ui.Space(ui.L12),
					form.Auto[TfObj](form.AutoOptions{}, state),
					ui.Space(ui.L32),
					ui.HStack(
						ui.Text("Gewählter Zeitraum:").Font(ui.Font{Weight: ui.DisplayAndLabelFontWeight}),
						ui.Text(date),
					).Alignment(ui.Leading).Gap(ui.L12).FullWidth(),
				).
					BackgroundColor(ui.M4).
					Border(ui.Border{}.Radius(ui.L12)).
					Padding(ui.Padding{}.All(ui.L12)).
					Frame(ui.Frame{MaxWidth: "60rem"}.FullWidth()),
			).FullWidth()
		})

	}).Run()
}
