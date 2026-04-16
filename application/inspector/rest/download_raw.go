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
	PathDownloadAsRaw = "/api/nago/v1/inspector/download/raw"
)

func NewDownloadAsRaw(p blob.Stores) http2.SubjectHandlerFunc {
	return newStoreHandler(p, func(w http.ResponseWriter, r *http.Request, subject auth.Subject, store blob.Store, item []string) {
		if len(item) != 1 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		key := item[0]
		optR, err := store.NewReader(r.Context(), key)
		if err != nil {
			slog.Error("cannot open blob for raw download", "key", key, "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if optR.IsNone() {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		rc := optR.Unwrap()
		defer rc.Close()

		// Lese bis zu 512 Bytes für die MIME-Type-Erkennung
		sniff := make([]byte, 512)
		n, err := io.ReadFull(rc, sniff)
		if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
			slog.Error("cannot read blob for mime sniffing", "key", key, "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		sniff = sniff[:n]

		contentType := http.DetectContentType(sniff)
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Disposition", `attachment; filename="`+key+`"`)

		// Bereits gelesene Bytes + Rest des Streams schreiben
		if _, err := w.Write(sniff); err != nil {
			slog.Error("cannot write sniff buffer for raw download", "key", key, "err", err)
			return
		}
		if _, err := io.Copy(w, rc); err != nil {
			slog.Error("cannot stream blob for raw download", "key", key, "err", err)
		}
	})
}
