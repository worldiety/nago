// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gollama

import (
	"fmt"
	"os"
	"path/filepath"
)

// resolveModel returns the on-disk path of the GGUF file for a catalog entry. It first looks in the configured
// search and storage folders; if the file is absent it is downloaded from HuggingFace into the storage folder.
// Downloads are serialised so two concurrent requests never fetch the same file twice.
func (e *engine) resolveModel(entry catalogEntry) (string, error) {
	storage := e.cfg.storageDir()
	target := filepath.Join(storage, entry.File)

	for _, p := range e.candidatePaths(entry) {
		if isRegularFile(p) {
			return p, nil
		}
	}

	e.dlMu.Lock()
	defer e.dlMu.Unlock()

	// Another goroutine may have completed the download while we waited for the lock.
	if isRegularFile(target) {
		return target, nil
	}

	if entry.HFRepo == "" {
		return "", fmt.Errorf("model %q not found in %v and no HuggingFace repository configured", entry.File, e.candidatePaths(entry))
	}

	if err := downloadModel(entry, storage, e.cfg.HFToken); err != nil {
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
