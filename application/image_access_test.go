// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application_test

import (
	"net/http"
	"testing"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/image"
	httpimage "go.wdy.de/nago/application/image/http"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/nagotest"
)

// The image endpoint asks the application with the subject of the request, which is anonymous without login.
func TestImageAccessOfTheConfigurator(t *testing.T) {
	type call struct {
		valid bool
		id    image.ID
	}
	calls := make(chan call, 1)

	base := nagotest.Serve(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.imageaccesstest")
		cfg.SetImageAccess(func(subject auth.Subject, id image.ID) bool {
			calls <- call{valid: subject != nil && subject.Valid(), id: id}
			return false
		})
	})

	res, err := http.Get(base + string(httpimage.URI("secret", image.FitCover, 64, 64)))
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("a denied image must look unknown, got %d", res.StatusCode)
	}

	select {
	case c := <-calls:
		if c.valid || c.id != "secret" {
			t.Fatalf("unexpected call %+v", c)
		}
	default:
		t.Fatal("the endpoint did not ask the application")
	}
}
