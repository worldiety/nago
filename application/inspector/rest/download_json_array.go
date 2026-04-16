// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package rest

import (
	"io"
	"log/slog"
	"net/http"

	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	http2 "go.wdy.de/nago/presentation/core/http"
)

const (
	PathDownloadAsJSONArray = "/api/nago/v1/inspector/download/json/array"
)

func NewDownloadAsJSONArray(p blob.Stores) http2.SubjectHandlerFunc {
	return newStoreHandler(p, func(w http.ResponseWriter, r *http.Request, subject auth.Subject, store blob.Store, items []string) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", `attachment; filename="export.json"`)

		_, _ = w.Write([]byte("["))

		first := true
		for _, key := range items {
			optR, err := store.NewReader(r.Context(), key)
			if err != nil {
				slog.Error("cannot open blob for JSON array download", "key", key, "err", err)
				continue
			}

			if optR.IsNone() {
				continue
			}

			if !first {
				_, _ = w.Write([]byte(","))
			}
			first = false

			rc := optR.Unwrap()
			if _, err := io.Copy(w, rc); err != nil {
				slog.Error("cannot stream blob for JSON array download", "key", key, "err", err)
			}
			_ = rc.Close()
		}

		if _, err := w.Write([]byte("]")); err != nil {
			slog.Error("cannot write closing bracket for JSON array download", "err", err)
		}
	})
}
