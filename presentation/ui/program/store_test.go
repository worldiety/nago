// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package program

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/presentation/core"
)

type sessionWindow struct {
	core.Window
	id session.ID
}

func (w sessionWindow) Session() session.UserSession { return fakeSession{id: w.id} }

type fakeSession struct {
	session.UserSession
	id session.ID
}

func (s fakeSession) ID() session.ID { return s.id }

// A component binds the URL of its bundle to the session of its window, so that a leaked hash grants no access.
func TestBundleIsBoundToTheSession(t *testing.T) {
	b := &gostrict.Bundle{Hash: "boundtestbundle", JS: "(function(){})"}
	uri := registerFor(sessionWindow{id: "alice"}, b)

	get := func(url, sessionID string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		SessionHandler(func(r *http.Request) string { return sessionID }).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, url, nil))
		return rec
	}

	rec := get(string(uri), "alice")
	if rec.Code != http.StatusOK {
		t.Fatalf("the session of the window cannot load its bundle: %d", rec.Code)
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "private, max-age=31536000, immutable" {
		t.Fatalf("unexpected Cache-Control %q", cc)
	}

	for _, req := range []struct{ url, session string }{
		{string(uri), "mallory"},                 // the URL in another session
		{string(uri), ""},                        // the URL without session
		{BundlePath + b.Hash + ".json", "alice"}, // the hash alone
	} {
		if rec := get(req.url, req.session); rec.Code != http.StatusNotFound {
			t.Fatalf("%s in session %q: expected 404, got %d", req.url, req.session, rec.Code)
		}
	}

	// a bundle of Register stays loadable by any session
	Register(b)
	if rec := get(BundlePath+b.Hash+".json", "mallory"); rec.Code != http.StatusOK {
		t.Fatalf("the public bundle cannot be loaded: %d", rec.Code)
	}
}
