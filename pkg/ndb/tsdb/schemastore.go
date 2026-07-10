// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const schemaFileName = "schema.json"

// loadSchema reads and validates the schema.json in dir. The bool is false if
// the file does not exist yet.
func loadSchema(dir string) (Schema, bool, error) {
	data, err := os.ReadFile(filepath.Join(dir, schemaFileName))
	if err != nil {
		if os.IsNotExist(err) {
			return Schema{}, false, nil
		}
		return Schema{}, false, fmt.Errorf("tsdb: read schema: %w", err)
	}
	var s Schema
	if err := json.Unmarshal(data, &s); err != nil {
		return Schema{}, false, fmt.Errorf("tsdb: parse schema: %w", err)
	}
	if err := s.validate(); err != nil {
		return Schema{}, false, err
	}
	return s, true, nil
}

// storeSchema writes schema.json atomically (temp file + rename), honoring the
// house rule that any non-append rewrite goes through a temp file.
func storeSchema(dir string, s Schema) error {
	if err := s.validate(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("tsdb: marshal schema: %w", err)
	}
	return atomicWrite(filepath.Join(dir, schemaFileName), data)
}

// atomicWrite writes data to a temp file in the same directory and renames it
// into place, so a reader never observes a partially written file.
func atomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("tsdb: create temp: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("tsdb: write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("tsdb: close temp: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("tsdb: rename temp: %w", err)
	}
	return nil
}
