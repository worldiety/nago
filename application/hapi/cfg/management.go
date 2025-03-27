// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfghapi

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/hapi"
	"go.wdy.de/nago/pkg/oas/v30"
	"go.wdy.de/nago/pkg/swagger"
	"log/slog"
	"net/http"
)

type Management struct {
	API *hapi.API
}

func Enable(cfg *application.Configurator) (Management, error) {
	management, ok := application.SystemServiceFor[Management](cfg, "")
	if ok {
		return management, nil
	}

	oapi := &oas.OpenAPI{
		Openapi: oas.Version,
		Info: oas.Info{
			Title:   cfg.Name(),
			Version: cfg.SemanticVersion(),
			Contact: &oas.Contact{
				Name:  "worldiety GmbH",
				URL:   "/",
				Email: "impressum@example.com",
			},
		},

		Paths: oas.Paths{},
	}

	management.API = hapi.NewAPI(hapi.Options{
		RegisterHandler: func(method, pattern string, handler http.HandlerFunc) {
			cfg.HandleMethod(method, pattern, handler)
		},
		OperationConfigured: func(op hapi.Operation, in hapi.Input, out hapi.Output) {
			item := &oas.PathItem{}

			switch op.Method {
			case http.MethodGet:
				item.Get = oasOpFrom(op, in, out)
			case http.MethodPost:
				item.Post = oasOpFrom(op, in, out)
			case http.MethodPut:
				item.Put = oasOpFrom(op, in, out)
			case http.MethodDelete:
				item.Delete = oasOpFrom(op, in, out)
			case http.MethodOptions:
				item.Options = oasOpFrom(op, in, out)
			case http.MethodHead:
				item.Head = oasOpFrom(op, in, out)
			case http.MethodPatch:
				item.Patch = oasOpFrom(op, in, out)
			case http.MethodTrace:
				item.Trace = oasOpFrom(op, in, out)
			default:
				slog.Error("unknown operation method", "method", op.Method)
			}
			oapi.Paths[op.Path] = item
		},
	})

	cfg.HandleFunc("/api/doc/spec.json", swagger.HandleOAS(oapi))

	cfg.AddSystemService("nago.api.hapi", management)
	slog.Info("installed user api management")

	return management, nil
}

func oasOpFrom(op hapi.Operation, in hapi.Input, out hapi.Output) *oas.Operation {
	o := &oas.Operation{
		Summary:     op.Summary,
		Description: op.Description,
		Parameters:  []oas.Parameter{},
		RequestBody: nil,
		Responses:   oas.Responses{},
		Deprecated:  op.Deprecated,
		Security:    nil,
	}

	in.DescribeInput(o)
	out.DescribeOutput(o)

	return o
}
