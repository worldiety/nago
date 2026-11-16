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
	"path/filepath"

	"go.wdy.de/nago/pkg/ndb"
)

const schemaFileName = "schema.json"

// loadSchema reads and validates the schema.json in dir through the pool. The bool is false if the file does not
// exist yet. A version larger than schemaVersion is errUnknownFormat: a later version of tsdb wrote it.
func loadSchema(fs *ndb.FilePool, dir string) (Schema, bool, error) {
	data, err := fs.ReadFile(filepath.Join(dir, schemaFileName))
	if err != nil {
		if isNotExist(err) {
			return Schema{}, false, nil
		}
		return Schema{}, false, fmt.Errorf("tsdb: read schema: %w", err)
	}
	var s Schema
	if err := json.Unmarshal(data, &s); err != nil {
		return Schema{}, false, fmt.Errorf("tsdb: parse schema: %w", err)
	}
	if s.Version > schemaVersion {
		return Schema{}, false, fmt.Errorf("%w: %s has version %d", errUnknownFormat, filepath.Join(dir, schemaFileName), s.Version)
	}
	if err := s.validate(); err != nil {
		return Schema{}, false, err
	}
	return s, true, nil
}

// storeSchema writes schema.json atomically (new file + rename), honoring the
// house rule that any non-append rewrite goes through a file of another name.
func storeSchema(fs *ndb.FilePool, dir string, s Schema) error {
	if err := s.validate(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("tsdb: marshal schema: %w", err)
	}
	path := filepath.Join(dir, schemaFileName)
	tmp := path + newSuffix
	if err := fs.WriteFile(tmp, data); err != nil {
		_ = fs.Remove(tmp)
		return fmt.Errorf("tsdb: write schema: %w", err)
	}
	if err := fs.Rename(tmp, path); err != nil {
		_ = fs.Remove(tmp)
		return fmt.Errorf("tsdb: rename schema: %w", err)
	}
	return nil
}
