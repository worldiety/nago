// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package rest

import (
	"archive/zip"
	"io"
	"log/slog"
	"net/http"

	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	http2 "go.wdy.de/nago/presentation/core/http"
)

const (
	PathDownloadAsZip = "/api/nago/v1/inspector/download/zip"
)

func NewDownloadAsZip(p blob.Stores) http2.SubjectHandlerFunc {
	return newStoreHandler(p, func(w http.ResponseWriter, r *http.Request, subject auth.Subject, store blob.Store, items []string) {
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", `attachment; filename="export.zip"`)

		zw := zip.NewWriter(w)

		for _, key := range items {
			optR, err := store.NewReader(r.Context(), key)
			if err != nil {
				slog.Error("cannot open blob for zip download", "key", key, "err", err)
				continue
			}

			if optR.IsNone() {
				continue
			}

			fh := &zip.FileHeader{
				Name:   EncodeZipKey(key),
				Method: zip.Deflate,
			}

			// populate ModTime without extra I/O if the store supports StatReader
			if info, err := blob.Stat(r.Context(), store, key); err == nil && info.IsSome() {
				if t := info.Unwrap().ModTime; !t.IsZero() {
					fh.Modified = t
				}
			}

			fw, err := zw.CreateHeader(fh)
			if err != nil {
				slog.Error("cannot create zip entry", "key", key, "err", err)
				_ = optR.Unwrap().Close()
				continue
			}

			rc := optR.Unwrap()
			if _, err := io.Copy(fw, rc); err != nil {
				slog.Error("cannot stream blob into zip", "key", key, "err", err)
			}
			_ = rc.Close()
		}

		// writes central directory + end-of-central-directory record
		if err := zw.Close(); err != nil {
			slog.Error("cannot finalize zip archive", "err", err)
		}
	})
}
