// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gollama

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// resolveModel returns the on-disk path of the GGUF file for a catalog entry. It first looks in the configured
// search and storage folders; if the file is absent it is downloaded from HuggingFace into the storage folder.
// A file is downloaded once, however many requests wait for it, see awaitDownload.
func (e *engine) resolveModel(ctx context.Context, entry catalogEntry) (string, error) {
	storage := e.cfg.storageDir()
	target := filepath.Join(storage, entry.File)

	for _, p := range e.candidatePaths(entry) {
		if isRegularFile(p) {
			return p, nil
		}
	}

	if entry.HFRepo == "" {
		return "", fmt.Errorf("model %q not found in %v and no HuggingFace repository configured", entry.File, e.candidatePaths(entry))
	}

	if err := e.awaitDownload(ctx, entry, storage); err != nil {
		return "", fmt.Errorf("download model %q: %w", entry.ID, err)
	}

	return target, nil
}

// candidatePaths lists the locations scanned for an existing model file, in priority order.
func (e *engine) candidatePaths(entry catalogEntry) []string {
	var paths []string
	if d := e.cfg.searchDir(); d != "" {
		paths = append(paths, filepath.Join(d, entry.File))
	}
	if d := e.cfg.storageDir(); d != "" {
		p := filepath.Join(d, entry.File)
		if len(paths) == 0 || paths[0] != p {
			paths = append(paths, p)
		}
	}
	return paths
}

// isRegularFile reports whether path exists and is a regular (non-directory) file.
func isRegularFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}
