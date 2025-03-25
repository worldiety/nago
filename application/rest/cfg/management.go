package cfgrest

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/pkg/swagger"
	"go.wdy.de/nago/pkg/swagger/oas"
	"log/slog"
)

type Management struct {
}

func Enable(cfg *application.Configurator) (Management, error) {
	management, ok := application.SystemServiceFor[Management](cfg, "")
	if ok {
		return management, nil
	}

	cfg.HandleFunc("/api/doc/spec.json", swagger.HandleOAS(oas.OpenAPI{
		Openapi: oas.OAS30,
		Info: oas.Info{
			Title:   "Nago",
			Version: "1.0",
			Contact: &oas.Contact{
				Name:  "worldiety GmbH",
				URL:   "/",
				Email: "impressum@example.com",
			},
		},

		Paths: oas.Paths{
			"/test": oas.PathItem{
				Get: &oas.Operation{
					Description: "Test Description",
					Summary:     "Test Summary",
					Parameters: []oas.Parameter{
						{
							Name:        "asdf",
							In:          oas.LocationQuery,
							Description: "Asdf",
						},
					},
				},
			},
		},
	}))

	cfg.AddSystemService("nago.api.rest", management)
	slog.Info("installed user api management")

	return management, nil
}
