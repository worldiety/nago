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
		OperationConfigured: func(op hapi.Operation) {
			item := &oas.PathItem{}

			switch op.Method {
			case http.MethodGet:
				item.Get = oasOpFrom(op)
			case http.MethodPost:
				item.Post = oasOpFrom(op)
			case http.MethodPut:
				item.Put = oasOpFrom(op)
			case http.MethodDelete:
				item.Delete = oasOpFrom(op)
			case http.MethodOptions:
				item.Options = oasOpFrom(op)
			case http.MethodHead:
				item.Head = oasOpFrom(op)
			case http.MethodPatch:
				item.Patch = oasOpFrom(op)
			case http.MethodTrace:
				item.Trace = oasOpFrom(op)
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

func oasOpFrom(op hapi.Operation) *oas.Operation {
	return &oas.Operation{
		Summary:     op.Summary,
		Description: op.Description,
		Parameters:  []oas.Parameter{},
		RequestBody: nil,
		Responses:   map[oas.HttpStatusOrDefault]oas.Ref{},
		Deprecated:  op.Deprecated,
		Security:    nil,
	}
}
