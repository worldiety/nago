package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/ui2"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.Component(".", func(wnd core.Window) core.View {

			return ui.HStack(
				defaultButtons(),
				ui.VLine(),
				customButtons(wnd),
			).Gap(ora.L16).Alignment(ora.Top).Frame(ora.Frame{}.FullWidth())

		})
	}).Run()
}

func defaultButtons() core.View {
	return ui.VStack(
		ui.Text("Standard Buttons").Font(ora.Title),

		ui.PrimaryButton(func() {
			fmt.Println("clicked the real primary")
		}).Title("primary button"),

		ui.PrimaryButton(nil).
			Title("primary with pre icon").
			PreIcon(icons.SpeakerWave),

		ui.PrimaryButton(nil).
			Title("primary with post icon").
			PostIcon(icons.SpeakerWave),

		ui.PrimaryButton(nil).
			PreIcon(icons.SpeakerWave),

		ui.Secondary(nil).Title("secondary button"),
		ui.Tertiary(nil).Title("tertiary button"),
	).Gap(ora.L16).
		Padding(ora.Padding{}.All(ora.L16))
}

func customButtons(wnd core.Window) core.View {
	colors := core.ColorSet[ora.Colors](wnd)
	return ui.VStack(
		ui.Text("Custom Buttons").Font(ora.Title),
		// we cannot use the variable "I0" because the function needs to calculate active and hover derivates
		ui.FilledButton(colors.I0, func() {
			fmt.Println("clicked a custom filled button")
		}).Title("fake primary button"),

		// hardcoded button, which does not react to color changes
		ui.FilledButton("#EF8A97", nil).TextColor("#ffffff").Title("arbitrary color"),
	).Gap(ora.L16).
		Padding(ora.Padding{}.All(ora.L16)) // graphical glitch: use some padding, custom buttons are 2dp larger due to emulated focus border, otherwise it gets clipped
}
