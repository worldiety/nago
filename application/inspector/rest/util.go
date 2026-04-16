// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package rest

import (
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"

	"go.wdy.de/nago/application/inspector"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	http2 "go.wdy.de/nago/presentation/core/http"
)

func newStoreHandler(p blob.Stores, fn func(w http.ResponseWriter, r *http.Request, subject auth.Subject, store blob.Store, items []string)) http2.SubjectHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, subject auth.Subject) {
		if !subject.HasPermission(inspector.PermDataInspector) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		query := r.URL.Query()
		optStore, err := p.Get(query.Get("store"))
		if err != nil {
			slog.Error("cannot get store", "store", query.Get("store"), "err", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if optStore.IsNone() {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		store := optStore.Unwrap()
		ids := DecodeQuery(query.Get("id"))
		if len(ids) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if len(ids) == 1 && ids[0] == "*" {
			keys, err := blob.Keys(store)
			if err != nil {
				slog.Error("cannot list keys download", "err", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			ids = keys
		}

		fn(w, r, subject, store, ids)
	}
}

func DecodeQuery(text string) []string {
	if len(text) == 0 {
		return nil
	}

	buf, err := base64.URLEncoding.DecodeString(text)
	if err != nil {
		return []string{text}
	}

	var tmp []string
	if err := json.Unmarshal(buf, &tmp); err != nil {
		return []string{text}
	}

	return tmp
}

func EncodeQuery(ids []string) string {
	buf, err := json.Marshal(ids)
	if err != nil {
		panic(err) // unreachable
	}

	return base64.URLEncoding.EncodeToString(buf)
}
