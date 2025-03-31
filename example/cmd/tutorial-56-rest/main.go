// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"context"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/hapi"
	cfghapi "go.wdy.de/nago/application/hapi/cfg"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/swagger"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
	"net/http"
	"time"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_56")
		cfg.SetSemanticVersion("0.1.2")
		cfg.SetName("Tutorial 56")

		cfg.Serve(vuejs.Dist())
		cfg.Serve(swagger.Dist())
		cfg.SetDecorator(cfg.NewScaffold().
			Decorator())

		option.MustZero(cfg.StandardSystems())

		std.Must(std.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))

		api := std.Must(cfghapi.Enable(cfg)).API
		std.Must(cfg.TokenManagement())
		configureMyAPI(api)

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			return VStack(Text("hello world")).
				Frame(Frame{}.MatchScreen())

		})
	}).
		Run()
}

func configureMyAPI(api *hapi.API) {

	type HelloResponse = hapi.JSON[string]

	hapi.Handle(api, hapi.Operation{Path: "/api/v1/hello"}, func(ctx context.Context, in *hapi.None) (*HelloResponse, error) {
		return &HelloResponse{
			Body: "hello world",
		}, nil
	})

	type HelloRequest struct {
		Name string
	}

	type HelloResponse2 struct {
		Msg string
	}

	hapi.Handle(api, hapi.Operation{Method: http.MethodPost, Path: "/api/v1/hello"}, func(ctx context.Context, in *hapi.JSON[HelloRequest]) (*hapi.JSON[HelloResponse2], error) {
		return &hapi.JSON[HelloResponse2]{
			Body: HelloResponse2{
				Msg: "hello " + in.Body.Name,
			},
		}, nil
	})
}
