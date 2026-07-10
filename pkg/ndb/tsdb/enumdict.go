// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
	"sync"

	"go.wdy.de/nago/pkg/ndb"
)

const enumDictName = "enum.dict"

// enum.dict on-disk entry:
//
//	| Len   4 | length of the string bytes (LE)
//	| CRC32 4 | IEEE over the string bytes
//	| Str   n | UTF-8 string
//
// The id of a variant is its zero-based ordinal in the file (append order).
// Ids are never reused. The whole dictionary is loaded into memory on open.
const enumEntryHdr = 8

// enumDict is an append-only string dictionary supporting flexible runtime
// extension. It maps string -> id and id -> string. Safe for concurrent use.
type enumDict struct {
	path string
	pool *ndb.FilePool

	mu     sync.RWMutex
	byStr  map[string]uint32
	byID   []string
	endOff int64
}

func openEnumDict(dir string, pool *ndb.FilePool) (*enumDict, error) {
	d := &enumDict{
		path:  filepath.Join(dir, enumDictName),
		pool:  pool,
		byStr: make(map[string]uint32),
	}
	if err := d.load(); err != nil {
		return nil, err
	}
	return d, nil
}

func (d *enumDict) load() error {
	fi, err := os.Stat(d.path)
	if err != nil {
		if os.IsNotExist(err) {
			d.endOff = 0
			return nil
		}
		return fmt.Errorf("tsdb: stat enum dict: %w", err)
	}
	size := fi.Size()
	if size == 0 {
		return nil
	}
	buf := make([]byte, size)
	if _, err := d.pool.ReadAt(d.path, buf, 0); err != nil {
		return fmt.Errorf("tsdb: read enum dict: %w", err)
	}
	off := 0
	for off+enumEntryHdr <= len(buf) {
		l := int(binary.LittleEndian.Uint32(buf[off:]))
		crc := binary.LittleEndian.Uint32(buf[off+4:])
		start := off + enumEntryHdr
		if l < 0 || start+l > len(buf) {
			break // truncated tail: stop, entry is incomplete
		}
		s := buf[start : start+l]
		if crc32.ChecksumIEEE(s) != crc {
			break // corruption: stop at first bad entry
		}
		id := uint32(len(d.byID))
		str := string(s)
		d.byID = append(d.byID, str)
		d.byStr[str] = id
		off = start + l
	}
	d.endOff = int64(off)
	return nil
}

// intern returns the id for s, appending a new entry if unseen. The append is a
// single pwrite of the entry frame; ids are assigned in append order.
func (d *enumDict) intern(s string) (uint32, error) {
	d.mu.RLock()
	if id, ok := d.byStr[s]; ok {
		d.mu.RUnlock()
		return id, nil
	}
	d.mu.RUnlock()

	d.mu.Lock()
	defer d.mu.Unlock()
	if id, ok := d.byStr[s]; ok {
		return id, nil
	}
	sb := []byte(s)
	frame := make([]byte, enumEntryHdr+len(sb))
	binary.LittleEndian.PutUint32(frame[0:], uint32(len(sb)))
	binary.LittleEndian.PutUint32(frame[4:], crc32.ChecksumIEEE(sb))
	copy(frame[enumEntryHdr:], sb)
	if _, err := d.pool.WriteAt(d.path, frame, d.endOff); err != nil {
		return 0, fmt.Errorf("tsdb: append enum dict: %w", err)
	}
	id := uint32(len(d.byID))
	d.byID = append(d.byID, s)
	d.byStr[s] = id
	d.endOff += int64(len(frame))
	return id, nil
}

// lookup resolves an id to its string. ok is false for unknown ids.
func (d *enumDict) lookup(id uint32) (string, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if int(id) >= len(d.byID) {
		return "", false
	}
	return d.byID[id], true
}
