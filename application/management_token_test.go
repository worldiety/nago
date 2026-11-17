// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application_test

import (
	"testing"

	"go.wdy.de/nago/application"
	cfghapi "go.wdy.de/nago/application/hapi/cfg"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/stoplight"
)

// TestTokenPageAPIDocButton: the token page of the admin center shows the button for the OpenAPI documentation only,
// if the application registers /api/doc/spec.json (cfghapi) and serves a page /api/doc/index.html (pkg/stoplight);
// with only one of both, the button would open a page without document or a 404. The row with both is the counter
// row of the others, and on each page the button to create a token proves that the page itself is there.
func TestTokenPageAPIDocButton(t *testing.T) {
	const label = "Open API Dokumentation anzeigen"

	tests := []struct {
		name        string
		spec, index bool
		want        bool
	}{
		{"neither document nor page", false, false, false},
		{"only the document", true, false, false},
		{"only the page", false, true, false},
		{"document and page", true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tokens application.TokenManagement
			app := nagotest.New(t, func(cfg *application.Configurator) {
				var err error
				if tokens, err = cfg.TokenManagement(); err != nil {
					t.Fatal(err)
				}

				if tt.spec {
					if _, err := cfghapi.Enable(cfg); err != nil {
						t.Fatal(err)
					}
				}

				if tt.index {
					cfg.Serve(stoplight.Dist())
				}
			})

			win := app.Open(t, user.SU(), tokens.Pages.Tokens)
			create := win.FindAll(nagotest.Text("Access Token hinzufügen"))
			if create.Len() != 1 {
				t.Fatalf("premise: the token page is not there, %d buttons to create a token", create.Len())
			}

			if got := win.FindAll(nagotest.Label(label)).Len() > 0; got != tt.want {
				t.Errorf("button %q shown = %v, want %v", label, got, tt.want)
			}
		})
	}
}
