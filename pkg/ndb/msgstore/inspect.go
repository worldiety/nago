// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package msgstore

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"go.wdy.de/nago/pkg/ndb"
)

// TypeStat is cheap, metadata-only information about one event type. All fields
// are derived from the on-disk segment listing (directory entries and their
// sizes plus the seq range encoded in segment file names) without scanning any
// message payload, so it is safe to call over event types holding millions of
// messages.
type TypeStat struct {
	// Type is the event type identifier.
	Type ndb.TypeID
	// Segments is the number of segment files for this type.
	Segments int
	// Bytes is the total on-disk size of all segments (including headers).
	Bytes int64
	// MinSeq is the smallest sequence id present (0 if the type is empty). It is
	// the lower bound to start a Replay from.
	MinSeq ndb.Seq
	// MaxSeq is the largest sequence id known from finalized segments. For a
	// pending (open) segment the true max is not encoded in the file name; use
	// the global Replay upper bound or Get to reach the newest message.
	MaxSeq ndb.Seq
	// HasPending indicates an open (not yet finalized) segment exists, so more
	// messages beyond MaxSeq may be present.
	HasPending bool
}

// Types returns the event types present on disk, sorted ascending. It performs
// a single directory read and does not open or scan any segment, so it is cheap
// regardless of how many messages are stored.
func (db *DB) Types() ([]ndb.TypeID, error) {
	eventsDir := filepath.Join(db.dir, "events")
	entries, err := os.ReadDir(eventsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("msgstore: list types: %w", err)
	}
	var out []ndb.TypeID
	for _, e := range entries {
		if !e.IsDir() || !ndb.ValidTypeID(e.Name()) {
			continue
		}
		out = append(out, ndb.TypeID(e.Name()))
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out, nil
}

// TypeStat returns metadata-only statistics for one event type. It reads the
// segment directory (and file sizes) but never scans message payloads, so it is
// O(segments), not O(messages). A type with no segments yields a zero-value
// TypeStat with the given Type and no error.
func (db *DB) TypeStat(typeID ndb.TypeID) (TypeStat, error) {
	dir := filepath.Join(db.dir, "events", string(typeID))
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return TypeStat{Type: typeID}, nil
		}
		return TypeStat{}, fmt.Errorf("msgstore: stat type %q: %w", typeID, err)
	}

	stat := TypeStat{Type: typeID}
	var haveMin bool
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		minSeq, maxSeq, pending, perr := parseSegmentName(e.Name())
		if perr != nil {
			continue // skip non-segment / unrecognized files
		}
		stat.Segments++
		if info, ierr := e.Info(); ierr == nil {
			stat.Bytes += info.Size()
		}
		if !haveMin || minSeq < uint64(stat.MinSeq) {
			stat.MinSeq = ndb.Seq(minSeq)
			haveMin = true
		}
		if pending {
			stat.HasPending = true
			if ndb.Seq(minSeq) > stat.MaxSeq {
				stat.MaxSeq = ndb.Seq(minSeq)
			}
		} else if ndb.Seq(maxSeq) > stat.MaxSeq {
			stat.MaxSeq = ndb.Seq(maxSeq)
		}
	}
	return stat, nil
}

// CountType returns the exact number of live (non-tombstone) messages of one
// event type. Unlike TypeStat, this replays the whole type, so it is O(messages)
// — call it deliberately, not on a hot path. A missing or empty type yields 0.
func (db *DB) CountType(typeID ndb.TypeID) (int64, error) {
	var n int64
	for _, msg := range db.Replay([]ndb.TypeID{typeID}, 0, ndb.Seq(^uint64(0))) {
		if msg.IsTombstone() {
			continue
		}
		n++
	}
	return n, nil
}
