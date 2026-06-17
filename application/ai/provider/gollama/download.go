// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gollama

import (
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"go.wdy.de/nago/pkg/xhttp"
)

// hfBaseURL is the HuggingFace endpoint that resolves a repository file to its (possibly CDN-redirected) raw
// download.
const hfBaseURL = "https://huggingface.co"

// downloadTimeout bounds a single model download. GGUF files can be several gigabytes, so this is generous.
const downloadTimeout = 24 * time.Hour

// downloadModel fetches the catalog entry's GGUF file from HuggingFace into storageDir. The body is streamed to
// a temporary file in the same directory and atomically renamed on success, so an interrupted download never
// leaves a partial file under the final name. An optional token authenticates gated/private repositories.
func downloadModel(entry catalogEntry, storageDir, token string) error {
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		return fmt.Errorf("create storage dir: %w", err)
	}

	target := filepath.Join(storageDir, entry.File)
	rawURL := fmt.Sprintf("%s/%s/resolve/main/%s", hfBaseURL, entry.HFRepo, url.PathEscape(entry.hfFile()))

	tmp, err := os.CreateTemp(storageDir, "."+entry.File+".*.part")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()

	// Best-effort cleanup: removing the temp file is a no-op once it has been renamed away.
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
	}()

	slog.Info("downloading gguf model", "model", entry.ID, "url", rawURL, "target", target)

	req := xhttp.NewRequest().
		URL(rawURL).
		Query("download", "true").
		Assert2xx(true).
		Timeout(downloadTimeout).
		To(func(r io.Reader) error {
			_, err := io.Copy(tmp, r)
			return err
		})
	if token != "" {
		req = req.BearerAuthentication(token)
	}

	if err := req.Get(); err != nil {
		return err
	}

	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("flush download: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close download: %w", err)
	}

	if err := os.Rename(tmpName, target); err != nil {
		return fmt.Errorf("finalize download: %w", err)
	}

	slog.Info("downloaded gguf model", "model", entry.ID, "target", target)
	return nil
}
