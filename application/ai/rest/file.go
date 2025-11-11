// Copyright (c) 2025 worldiety GmbH
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

	"go.wdy.de/nago/application/ai"
	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
)

const Endpoint = "/api/nago/v1/ai/file"

func URL(provider provider.ID, file file.ID) core.URI {
	return core.URI(Endpoint + "?id=" + string(file) + "&p=" + string(provider))
}

func NewFileEndpoint(findProvider ai.FindProviderByID) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		values := r.URL.Query()
		p := provider.ID(values.Get("p"))
		optProv, err := findProvider(user.SU(), p)
		if err != nil {
			slog.Error("failed to find provider", "id", p, "err", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if optProv.IsNone() {
			slog.Error("failed to find provider", "id", p)
			http.Error(w, "", http.StatusNotFound)
			return
		}

		prov := optProv.Unwrap()
		if prov.Files().IsNone() {
			slog.Error("provider does not support files", "id", p)
			http.Error(w, "", http.StatusNotFound)
			return
		}

		id := values.Get("id")
		// TODO this endpoint is currently unauthenticated. we must not rely on "secret" ids, the content may be important. However, you have sent your secret stuff anyway to some untrusty AI and therefore lost it already
		optReader, err := prov.Files().Unwrap().Get(user.SU(), file.ID(id))
		if err != nil {
			slog.Error("failed to open ai file", "id", id, "err", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if optReader.IsNone() {
			http.Error(w, "", http.StatusNotFound)
			return
		}

		reader := optReader.Unwrap()
		defer reader.Close()

		if _, err := io.Copy(w, reader); err != nil {
			slog.Error("failed to write ai file into http", "id", id, "err", err.Error())
		}
	}
}
