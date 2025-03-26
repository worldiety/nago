package main

import (
	"context"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/hapi"
	cfghapi "go.wdy.de/nago/application/hapi/cfg"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/swagger"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_56")
		cfg.SetSemanticVersion("0.1.2")
		cfg.SetName("Tutorial 56")

		cfg.Serve(vuejs.Dist())
		cfg.Serve(swagger.Dist())

		api := std.Must(cfghapi.Enable(cfg)).API
		configureMyAPI(api)

		cfg.RootView(".", func(wnd core.Window) core.View {
			return VStack(Text("hello world")).
				Frame(Frame{}.MatchScreen())

		})
	}).
		Run()
}

func configureMyAPI(api *hapi.API) {

	type HelloParams struct {
	}

	type HelloResponse struct {
		Body string
	}

	hapi.Handle(api, hapi.Operation{Path: "/api/v1/hello"}, func(ctx context.Context, in *HelloParams) (*HelloResponse, error) {
		return &HelloResponse{
			Body: "Hello world",
		}, nil
	})
}
