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
	"hash/maphash"
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
// Ids are never reused. The file is only ever appended to. The whole dictionary is loaded into memory when it is
// needed first, by a write or a read of the column, not when the column is opened.
const enumEntryHdr = 8

var dictSeed = maphash.MakeSeed()

// enumDict is an append-only string dictionary supporting flexible runtime
// extension. It maps string -> id and id -> string. Safe for concurrent use.
//
// byID holds the strings, index finds an id by its string: an open-addressing table of id+1 (0 is empty) at the hash
// of the string, at most half full, so that an entry takes 16 bytes in byID and 8 to 16 in index besides its string.
type enumDict struct {
	path string
	fs   *ndb.FilePool
	size int64 // what the listing of the column found of the file: 0 for none, -1 for unknown

	mu     sync.RWMutex
	loaded bool
	byID   []string
	index  []uint32
	endOff int64
	gone   bool // the column was dropped; a later file of that name belongs to a new column
}

func newEnumDict(dir string, fs *ndb.FilePool, size int64) *enumDict {
	return &enumDict{path: filepath.Join(dir, enumDictName), fs: fs, size: size}
}

// ensure loads the dictionary, if that did not happen yet. Caller holds d.mu for writing.
func (d *enumDict) ensure() error {
	if d.loaded {
		return nil
	}
	if d.size != 0 {
		if err := d.load(); err != nil {
			return err
		}
	}
	d.loaded = true
	return nil
}

func (d *enumDict) load() error {
	buf, err := d.fs.ReadFile(d.path)
	if err != nil {
		if isNotExist(err) {
			return nil
		}
		return fmt.Errorf("tsdb: read enum dict: %w", err)
	}
	all := string(buf) // one allocation, every entry is a part of it
	off := 0
	for off+enumEntryHdr <= len(buf) {
		l := int(binary.LittleEndian.Uint32(buf[off:]))
		crc := binary.LittleEndian.Uint32(buf[off+4:])
		start := off + enumEntryHdr
		if l < 0 || start+l > len(buf) {
			break // truncated tail: stop, entry is incomplete
		}
		if crc32.ChecksumIEEE(buf[start:start+l]) != crc {
			break // corruption: stop at first bad entry
		}
		d.add(all[start : start+l])
		off = start + l
	}
	d.endOff = int64(off)
	return nil
}

// find returns the id of s. Caller holds d.mu.
func (d *enumDict) find(s string) (uint32, bool) {
	if len(d.index) == 0 {
		return 0, false
	}
	mask := uint64(len(d.index) - 1)
	for i := maphash.String(dictSeed, s) & mask; ; i = (i + 1) & mask {
		v := d.index[i]
		if v == 0 {
			return 0, false
		}
		if d.byID[v-1] == s {
			return v - 1, true
		}
	}
}

// add gives s the next id. Caller holds d.mu for writing.
func (d *enumDict) add(s string) uint32 {
	id := uint32(len(d.byID))
	d.byID = append(d.byID, s)
	if 2*len(d.byID) > len(d.index) {
		n := max(8, 2*len(d.index))
		d.index = make([]uint32, n)
		for i, str := range d.byID {
			d.place(str, uint32(i))
		}
		return id
	}
	d.place(s, id)
	return id
}

func (d *enumDict) place(s string, id uint32) {
	mask := uint64(len(d.index) - 1)
	i := maphash.String(dictSeed, s) & mask
	for d.index[i] != 0 {
		i = (i + 1) & mask
	}
	d.index[i] = id + 1
}

// intern returns the id for s, appending a new entry if unseen. The append is a
// single pwrite of the entry frame; ids are assigned in append order.
func (d *enumDict) intern(s string) (uint32, error) {
	d.mu.RLock()
	if d.loaded {
		if id, ok := d.find(s); ok {
			d.mu.RUnlock()
			return id, nil
		}
	}
	d.mu.RUnlock()

	d.mu.Lock()
	defer d.mu.Unlock()
	if d.gone {
		return 0, errColumnGone
	}
	if err := d.ensure(); err != nil {
		return 0, err
	}
	if id, ok := d.find(s); ok {
		return id, nil
	}
	frame := make([]byte, enumEntryHdr+len(s))
	binary.LittleEndian.PutUint32(frame[0:], uint32(len(s)))
	binary.LittleEndian.PutUint32(frame[4:], crc32.ChecksumIEEE([]byte(s)))
	copy(frame[enumEntryHdr:], s)
	if _, err := d.fs.WriteAt(d.path, frame, d.endOff); err != nil {
		return 0, fmt.Errorf("tsdb: append enum dict: %w", err)
	}
	d.endOff += int64(len(frame))
	d.size = d.endOff
	return d.add(s), nil
}

// abandon makes each later intern fail, so that a handle of a dropped column can not write to the dictionary of a new
// column of the same name.
func (d *enumDict) abandon() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.gone = true
}

// lookup resolves an id to its string. ok is false for unknown ids.
func (d *enumDict) lookup(id uint32) (string, bool) {
	d.mu.RLock()
	if !d.loaded {
		d.mu.RUnlock()
		d.mu.Lock()
		err := d.ensure()
		d.mu.Unlock()
		if err != nil {
			return "", false
		}
		d.mu.RLock()
	}
	defer d.mu.RUnlock()
	if int(id) >= len(d.byID) {
		return "", false
	}
	return d.byID[id], true
}

// loadErr loads the dictionary, if that did not happen yet, for a read that can report an error of it.
func (d *enumDict) loadErr() error {
	d.mu.RLock()
	loaded := d.loaded
	d.mu.RUnlock()
	if loaded {
		return nil
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.ensure()
}
