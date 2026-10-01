// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgchatbot

import (
	"testing"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/nagotest"
)

// TestEnableIsIdempotent ensures that a second Enable short-circuits via the context lookup instead of
// registering the RootViews again, which would panic on the duplicate id.
func TestEnableIsIdempotent(t *testing.T) {
	nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.chatbottest")

		first, err := Enable(cfg)
		if err != nil {
			t.Fatal(err)
		}

		second, err := Enable(cfg)
		if err != nil {
			t.Fatal(err)
		}

		if first.Pages != second.Pages {
			t.Fatalf("expected the same pages, got %+v and %+v", first.Pages, second.Pages)
		}
	})
}
