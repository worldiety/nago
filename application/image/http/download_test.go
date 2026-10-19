// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package httpimage

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.wdy.de/nago/application/image"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std"
)

func loadFake(loaded *[]image.ID) image.LoadBestFit {
	return func(_ permission.Auditable, id image.ID, _ image.ObjectFit, _, _ int) (std.Option[io.ReadCloser], error) {
		*loaded = append(*loaded, id)
		return std.Some[io.ReadCloser](io.NopCloser(strings.NewReader("png"))), nil
	}
}

// An image may belong to a user, so no shared cache may keep it.
func TestImageIsCachedPrivately(t *testing.T) {
	var loaded []image.ID
	rec := httptest.NewRecorder()
	NewHandler(loadFake(&loaded)).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, string(URI("a", image.FitCover, 64, 64)), nil))
	if rec.Code != http.StatusOK || rec.Body.String() != "png" {
		t.Fatalf("unexpected response %d %q", rec.Code, rec.Body.String())
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "private, max-age=31536000" {
		t.Fatalf("unexpected Cache-Control %q", cc)
	}
}

// The application decides by the subject, whether an image may be loaded, and a denied image looks like an unknown one.
func TestImageAccessIsChecked(t *testing.T) {
	var loaded []image.ID
	handler := NewSubjectHandler(loadFake(&loaded), func(subject auth.Subject, id image.ID) bool {
		return id == "public"
	})

	for id, want := range map[image.ID]int{"public": http.StatusOK, "private": http.StatusNotFound} {
		rec := httptest.NewRecorder()
		handler(rec, httptest.NewRequest(http.MethodGet, string(URI(id, image.FitCover, 64, 64)), nil), nil)
		if rec.Code != want {
			t.Fatalf("%s: expected %d, got %d", id, want, rec.Code)
		}
	}

	if len(loaded) != 1 || loaded[0] != "public" {
		t.Fatalf("a denied image has been loaded: %v", loaded)
	}
}
