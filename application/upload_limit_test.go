// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

// The frontend may ignore the MaxBytes of an import, so the server must reject a larger upload.
func TestUploadIsLimitedByTheImport(t *testing.T) {
	const maxBytes = 1024
	scopeID := proto.ScopeID(strings.Repeat("s", 32))

	upload := func(size int) (int, []core.File) {
		var received []core.File
		handler := uploadHandler(func(id proto.ScopeID, uploadID string) (core.ImportFilesOptions, bool) {
			if id != scopeID || uploadID != "pick" {
				return core.ImportFilesOptions{}, false
			}

			return core.ImportFilesOptions{
				ID:       uploadID,
				MaxBytes: maxBytes,
				OnCompletion: func(files []core.File) {
					received = files
				},
			}, true
		})

		var body bytes.Buffer
		form := multipart.NewWriter(&body)
		part, err := form.CreateFormFile("a.bin", "a.bin")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write(bytes.Repeat([]byte{'x'}, size)); err != nil {
			t.Fatal(err)
		}
		if err := form.Close(); err != nil {
			t.Fatal(err)
		}

		r := httptest.NewRequest(http.MethodPost, "/api/ora/v1/upload", &body)
		r.Header.Set("Content-Type", form.FormDataContentType())
		r.Header.Set("x-scope", string(scopeID))
		r.Header.Set("x-receiver", "pick")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, r)
		return rec.Code, received
	}

	if code, files := upload(maxBytes); code != http.StatusOK || len(files) != 1 {
		t.Fatalf("an upload within the limit failed: %d, %d files", code, len(files))
	}

	if code, files := upload(maxBytes + multipartOverhead + 1); code != http.StatusRequestEntityTooLarge || files != nil {
		t.Fatalf("an upload above the limit was not rejected: %d, %d files", code, len(files))
	}
}
