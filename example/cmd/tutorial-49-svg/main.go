package main

import (
	_ "embed"
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"strings"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.SetDecorator(cfg.NewScaffold().
			Logo(ui.Image().Embed(heroSolid.AcademicCap).Frame(ui.Frame{}.Size(ui.L96, ui.L96))).
			Decorator())

		cfg.RootView(".", cfg.DecorateRootView(func(wnd core.Window) core.View {

			return ui.VStack(
				ClickableSVG{
					Actions: [4]func(){
						func() {
							fmt.Println("Hello world 0")
						},
						func() {
							fmt.Println("Hello world 1")
						},
						func() {
							fmt.Println("Hello world 2")
						},
						func() {
							fmt.Println("Hello world 3")
						},
					},
				},
			).Gap(ui.L8).Frame(ui.Frame{}.MatchScreen())
		}))

	}).Run()
}

//go:embed svg.svg
var img string

type ClickableSVG struct {
	Actions [4]func()
}

func (c ClickableSVG) Render(ctx core.RenderContext) core.RenderNode {
	// Important: place aria-valuenow attributes with the according callback pointer. We must misuse properly defined attributes, to pass them properly up the toolchain.
	tmp := img
	for i := range 4 {
		placeholder := fmt.Sprintf("VAR%d", i)
		call := fmt.Sprintf("%d", ctx.MountCallback(c.Actions[i]))
		tmp = strings.Replace(tmp, placeholder, call, 1)
	}

	return ui.Image().Embed(core.SVG(tmp)).Frame(ui.Frame{}.Size(ui.L480, ui.L480)).Render(ctx)
}
