// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	ui "go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"log/slog"
	"net/http"
	"time"
)

type rawEndpoint struct {
	method  string // if method is empty, the endpoint will respond to any method
	pattern string
	handler http.HandlerFunc
}

type Resource interface {
	configureResource(c *Configurator) proto.URI
}

type StaticBytes []byte

func (r StaticBytes) configureResource(c *Configurator) proto.URI {
	sum := sha256.Sum256(r)
	token := hex.EncodeToString(sum[:])
	pattern := fmt.Sprintf("/api/ora/v1/static/%s", token)
	mimeType := magicMimeType(r)
	c.rawEndpoint = append(c.rawEndpoint, rawEndpoint{
		pattern: pattern,
		handler: func(writer http.ResponseWriter, request *http.Request) {
			// enable aggressive caching, because we have a stable resource identifier based on a hash sum
			writer.Header().Set("Cache-Control", "public, max-age=31536000")
			writer.Header().Set("Content-Type", mimeType)
			expires := time.Now().Add(365 * 24 * time.Hour)
			writer.Header().Set("Expires", expires.Format(http.TimeFormat))

			if _, err := writer.Write(r); err != nil {
				slog.Error("failed to write response", "err", err)
			}
		},
	})

	return proto.URI(pattern)
}

func magicMimeType(buf []byte) string {
	if bytes.Contains(buf[:min(len(buf), 1024)], []byte("<svg")) {
		return "image/svg+xml"
	}

	return "application/octet-stream"
}

// Resource registers the given resource. It will likely result in an additional endpoint which looks like
// /api/ora/v1/resource/<some random identifier>
func (c *Configurator) Resource(r Resource) ui.URI {
	return ui.URI(r.configureResource(c))
}
