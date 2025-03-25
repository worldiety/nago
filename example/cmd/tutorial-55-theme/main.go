package main

import (
	"fmt"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/theme"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/form"
	"go.wdy.de/nago/web/vuejs"
	"time"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_55")
		cfg.Serve(vuejs.Dist())

		cfg.SetDecorator(cfg.NewScaffold().Decorator())
		option.MustZero(cfg.StandardSystems())

		myTheme := option.Must(cfg.ThemeManagement())

		if !option.Must(myTheme.UseCases.HasColors(user.SU())) {
			myBaseColors := theme.BaseColors{
				Main:        "#f12af7",
				Interactive: "#2af4f7",
				Accent:      "#5af72a",
			}

			myDark := myTheme.UseCases.Calculations.DarkMode(myBaseColors)
			myLight := myTheme.UseCases.Calculations.LightMode(myBaseColors)

			option.MustZero(myTheme.UseCases.UpdateColors(user.SU(), theme.Colors{
				Dark:  myDark,
				Light: myLight,
			}))
		}

		std.Must(std.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {

			colorState := core.AutoState[theme.BaseColors](wnd).Init(func() theme.BaseColors {
				actual := option.Must(myTheme.UseCases.ReadColors(user.SU()))
				return theme.BaseColors{
					Main:        actual.Light.M0,
					Interactive: actual.Light.I0,
					Accent:      actual.Light.A0,
				}
			})

			return ui.VStack(
				ui.Text("hello theme"),
				form.Auto[theme.BaseColors](form.AutoOptions{}, colorState).Frame(ui.Frame{MaxWidth: ui.L560, Width: ui.Full}),

				ui.PrimaryButton(func() {
					option.MustZero(myTheme.UseCases.ResetColors(user.SU()))
					fmt.Println("restart server to apply the effect")
				}).Title("Reset Colors"),

				ui.PrimaryButton(func() {
					option.MustZero(myTheme.UseCases.UpdateColors(user.SU(), theme.Colors{
						Dark:  myTheme.UseCases.Calculations.DarkMode(colorState.Get()),
						Light: myTheme.UseCases.Calculations.LightMode(colorState.Get()),
					}))

					wnd.Navigation().Reload()
				}).Title("Apply User Colors"),
			).Gap(ui.L16).Frame(ui.Frame{}.MatchScreen())
		})

	}).Run()
}
