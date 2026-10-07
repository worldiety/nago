// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// A browser sends a header value as Latin-1, so the frontend percent-encodes the import id of an upload.
func TestUploadReceiverDecodesTheImportID(t *testing.T) {
	for raw, want := range map[string]string{
		"Datei%20w%C3%A4hlen": "Datei wählen",
		"auto-1":              "auto-1",
		"label%2520encoded":   "label%20encoded", // an id which has been percent-encoded by the app itself
		"broken%zz":           "broken%zz",
	} {
		r := httptest.NewRequest(http.MethodPost, "/api/ora/v1/upload", nil)
		r.Header.Set("x-receiver", raw)
		if got := uploadReceiver(r); got != want {
			t.Errorf("%q: got %q, want %q", raw, got, want)
		}
	}
}
