// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfghapi

import (
	"encoding/json"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/hapi"
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/application/theme"
	"go.wdy.de/nago/pkg/oas/v31"
	"go.wdy.de/nago/presentation/core"
	"log/slog"
	"net/http"
)

type Management struct {
	API *hapi.API
}

func Enable(cfg *application.Configurator) (Management, error) {
	management, ok := core.FromContext[Management](cfg.Context(), "")
	if ok {
		return management, nil
	}

	sets, err := cfg.SettingsManagement()
	if err != nil {
		return Management{}, err
	}

	cfgTheme := settings.ReadGlobal[theme.Settings](sets.UseCases.LoadGlobal)

	oapi := &oas.OpenAPI{
		Openapi: oas.Version,
		Info: oas.Info{
			Title:   cfg.Name(),
			Version: cfg.SemanticVersion() + " (" + cfg.VCSVersion() + ")",
			Contact: &oas.Contact{
				Name:  cfgTheme.ProviderName,
				URL:   cfgTheme.APIPage,
				Email: cfgTheme.APIContact,
			},
		},

		Paths: oas.Paths{},
	}

	management.API = hapi.NewAPI(oapi, hapi.Options{
		RegisterHandler: func(method, pattern string, handler http.HandlerFunc) {
			cfg.HandleMethod(method, pattern, handler)
		},
	})

	cfg.HandleFunc("/api/doc/spec.json", handleOAS(oapi))

	cfg.AddContextValue(core.ContextValue("nago.api.hapi", management))
	slog.Info("installed user api management")

	return management, nil
}

func handleOAS(spec *oas.OpenAPI) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		buf, err := json.MarshalIndent(spec, "  ", " ")
		if err != nil {
			slog.Error("failed to marshal oas spec", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if _, err := w.Write(buf); err != nil {
			slog.Error("failed to write oas spec", "err", err)
			return
		}
	}
}
