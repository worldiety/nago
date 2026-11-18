// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"go.wdy.de/nago/pkg/ndb"
)

// values is the one place, through which tsdb writes a value of a column into a file of the column (R-DEC-TS-NAGO-DICT):
// a continuation record (Sync, ReleaseIdle), a block (a seal, also by Close, Flush and the split policy), a record of
// the WAL (a correction or deletion out of order), the records of the open block written anew when a column opens, an
// old WAL rewritten in the format of this version, what a compaction writes (its chunks, its WAL) and what it moves
// behind its WAL. Before such a write, the dictionary of an enum column is on the disk with every entry it holds, with
// its entry in the directory for a new file (enumDict.secure), so that no file of the column, which counts at the next
// open, holds an id, whose text a power failure could take. A staged write goes into a file of a compaction, which
// counts only after its commit: there the dictionary goes into the SyncAll of the compaction instead (compaction.dict),
// which runs before the commit. Writes without a value (a header, a slot, a cut, a rename) do not go through here.
type values struct {
	pool *ndb.FilePool
	dict *enumDict // nil for a column without a dictionary
}

// write writes b at off of the file at path, or, with off < 0, the whole file. staged is the compaction, whose file
// path is, or nil.
func (v *values) write(path string, b []byte, off int64, staged *compaction) error {
	if staged != nil {
		staged.dict = v.dict
	} else if err := v.dict.secure(); err != nil {
		return err
	}
	if off < 0 {
		return v.pool.WriteFile(path, b)
	}
	_, err := v.pool.WriteAt(path, b, off)
	return err
}
