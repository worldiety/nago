// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package stoplight

import (
	"embed"
	"encoding/json"
	"go.wdy.de/nago/pkg/oas/v31"
	"log/slog"
	"net/http"
)

//go:embed api/doc
var files embed.FS

func HandleOAS(spec *oas.OpenAPI) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		buf, err := json.Marshal(spec)
		if err != nil {
			slog.Error("failed to marshal oas spec: %v", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if _, err := w.Write(buf); err != nil {
			slog.Error("failed to write oas spec: %v", "err", err)
			return
		}
	}
}
