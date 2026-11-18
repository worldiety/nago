// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"hash/maphash"
	"math"
	"path/filepath"
	"slices"
	"sync"

	"go.wdy.de/nago/pkg/ndb"
)

const enumDictName = "enum.dict"

// enum.dict, the dictionary of an enum column, in the format this version writes. It begins with a header
//
//	| Magic    4 | "NAGO"
//	| Format   4 | "NTSD" Nago Time Series Dictionary
//	| Version  1 | dictVersion
//	| CRC32    4 | IEEE over Magic .. Version
//
// and goes on with entries, each appended behind the one before:
//
//	| Marker   4 | dictMarker (LE)
//	| Kind     1 | dictText, dictLost or dictReserved
//	| ID       4 | text: the id of the text; lost and reserved: the first of their ids (LE)
//	| Len      4 | text: the length of the text; lost and reserved: the number of their ids (LE)
//	| Salt     8 | text: the key of the mask of the text, random for each entry (LE); lost and reserved: 0
//	| CRC32    4 | IEEE over Kind .. Salt and the text as it is stored, masked
//	| Text     n | text only: the UTF-8 text, masked with the key stream of Salt (maskStream at offset 0, chunk.go)
//
// Header, marker and length of an entry stand unmasked; no text, which a caller writes, stands in the file as written,
// like the bodies of the blocks of a chunk: the bytes of a text are those of the key stream of a salt, which the writer
// does not know, so that the search behind a damaged place (nextDictEntry) can not take a written text for an entry
// (R-DEC-TS-NAGO-DICT, answer „Texte maskieren“). The salt lies in each entry, not in the header of the file, so that a
// damaged salt costs only its entry, whose CRC covers it. The CRC covers the kind, the id, the length and the salt, and
// a run of zero bytes is no entry, because the marker is not zero. Ids are explicit and only grow in the order of the
// file, so that an entry behind a damaged one keeps its id, and the ids a damaged region held lie between those of the
// entries around it. A lost entry holds the ids of a damaged region of a dictionary,
// which nago rewrote from the format before N23 (convertDict); it is reported like a damaged entry at every load. A
// reserved entry holds ids, which an entry torn by a crash may have had; it is silent. No new value gets one of their
// ids. A point, whose id has no text, is skipped by a read and reported at the place of its id (missingPlace).
//
// nago tells this format from the one before N23 by the magic and the format of the header: a file is in this format,
// if either of both is in place, so that one damaged byte there costs no entry and does not make the file look like
// one before N23. A damaged header is reported at offset 0 and read as dictVersion; a header with its CRC and a later
// version is errUnknownFormat, and nago then neither reads, writes nor rewrites the file.
//
// The format before N23 (8ae1daaee) has no header, its entries are | Len 4 | CRC32 4 over the text | Text n |, and the
// id of an entry is its ordinal in the file. Only convertDict reads it: the first load of such a dictionary rewrites it
// once into this format, with the same ids, and nago reads and writes nothing else of it.
//
// The whole dictionary is loaded into memory when it is needed first, by a write or a read of the column, not when the
// column is opened. A new text is appended at once, but it is not waited for; secure puts the dictionary on the disk
// before a value goes into a file of the column (values.write).
const (
	dictMagic     = "NAGO"
	dictFormat    = "NTSD"
	dictVersion   = byte(1)
	dictHeaderLen = 4 + 4 + 1 + 4
	dictMarker    = uint32(0x7A5EC7D1)
	dictEntryHdr  = 4 + 1 + 4 + 4 + 8 + 4

	dictText     = byte(0)
	dictLost     = byte(1)
	dictReserved = byte(2)

	// dictOldEntryHdr is the header of an entry in the format before N23: | Len 4 | CRC32 4 |.
	dictOldEntryHdr = 8
)

// the texts of the reports of a dictionary (R-DEC-TS-NAGO-DAMAGE)
const (
	reportDictEntry   = "tsdb: damaged entry of an enum dictionary skipped"
	reportDictMissing = "tsdb: enum values without their text skipped"
)

var dictMarkerBytes = binary.LittleEndian.AppendUint32(nil, dictMarker)

var dictSeed = maphash.MakeSeed()

// enumDict is an append-only string dictionary supporting flexible runtime
// extension. It maps string -> id and id -> string. Safe for concurrent use.
//
// byID holds the strings at their ids and known tells which ids have one; index finds an id by its string: an
// open-addressing table of id+1 (0 is empty) at the hash of the string, at most half full, so that an entry takes 16
// bytes in byID and 8 to 16 in index besides its string.
type enumDict struct {
	path   string
	fs     *ndb.FilePool
	size   int64 // what the listing of the column found of the file: 0 for none, -1 for unknown
	report func(path string, offset int64, what string)

	mu     sync.RWMutex
	loaded bool
	byID   []string
	known  []uint64
	texts  int // the number of ids with a text
	index  []uint32
	nextID uint32
	// header is set, when the file begins with a header; else the next append writes one at offset 0
	header bool
	// endOff is where the next entry goes; cutAt, if not -1, is where a rest, which a crash tore, begins: the next
	// append cuts the file there first, so that no rest of it is ever read as an entry or as a damaged place
	endOff int64
	cutAt  int64
	// places are the ids without a text and where they belong, sorted by their ids; tail is the place of every id
	// after them (missingPlace)
	places []dictPlace
	tail   int64
	gone   bool // the column was dropped; a later file of that name belongs to a new column

	// secMu serializes secure. The file is on the disk up to securedEnd; with dirToSync its entry in the directory may
	// not be.
	secMu      sync.Mutex
	securedEnd int64
	dirToSync  bool
}

// dictPlace is where the ids lo..hi-1, which have no text, belong: a damaged region, a lost or reserved entry or a gap.
type dictPlace struct {
	lo, hi uint32
	off    int64
}

func newEnumDict(dir string, fs *ndb.FilePool, size int64, report func(path string, offset int64, what string)) *enumDict {
	return &enumDict{path: filepath.Join(dir, enumDictName), fs: fs, size: size, report: report, cutAt: -1}
}

// ensure loads the dictionary, if that did not happen yet. Caller holds d.mu for writing.
func (d *enumDict) ensure() error {
	if d.loaded {
		return nil
	}
	d.reset()
	if d.size != 0 {
		if err := d.load(); err != nil {
			d.reset()
			return err
		}
	}
	d.loaded = true
	return nil
}

// reset empties the state, as for a missing file.
func (d *enumDict) reset() {
	d.byID, d.known, d.index, d.places, d.texts = nil, nil, nil, nil, 0
	d.nextID, d.header, d.endOff, d.cutAt, d.tail = 0, false, 0, -1, 0
	d.securedEnd, d.dirToSync = 0, true
}

func (d *enumDict) load() error {
	buf, err := d.fs.ReadFile(d.path)
	if err != nil {
		if isNotExist(err) {
			return nil
		}
		return fmt.Errorf("tsdb: read enum dict: %w", err)
	}
	if len(buf) == 0 {
		return nil
	}
	if !isDictFormat(buf) {
		return d.convert(buf)
	}
	if err := d.parse(buf); err != nil {
		return err
	}
	// what an earlier life appended may not be on the disk yet: the next value written syncs it once
	d.securedEnd, d.dirToSync = 0, true
	return nil
}

// isDictFormat reports whether buf, a non-empty dictionary, is in the format of this version: its magic or its format
// is in place.
func isDictFormat(buf []byte) bool {
	return len(buf) >= 4 && string(buf[0:4]) == dictMagic || len(buf) >= 8 && string(buf[4:8]) == dictFormat
}

// dictHeader is the header of a new dictionary.
func dictHeader() []byte {
	h := make([]byte, dictHeaderLen)
	copy(h[0:], dictMagic)
	copy(h[4:], dictFormat)
	h[8] = dictVersion
	binary.LittleEndian.PutUint32(h[9:], crc32.ChecksumIEEE(h[:9]))
	return h
}

// dictEntry is an intact entry of a dictionary.
type dictEntry struct {
	kind  byte
	id    uint32 // the id of the text or the first id
	count uint32 // the number of ids: 1 for a text
	text  string
	size  int
}

// appendDictEntry appends the entry of kind with id: for a text with text, masked with a new salt, for lost and
// reserved with count ids.
func appendDictEntry(dst []byte, kind byte, id uint32, count uint32, text string) []byte {
	var hdr [dictEntryHdr]byte
	binary.LittleEndian.PutUint32(hdr[0:], dictMarker)
	hdr[4] = kind
	binary.LittleEndian.PutUint32(hdr[5:], id)
	var salt uint64
	if kind == dictText {
		count = uint32(len(text))
		salt = newSalt()
	}
	binary.LittleEndian.PutUint32(hdr[9:], count)
	binary.LittleEndian.PutUint64(hdr[13:], salt)
	start := len(dst)
	dst = append(append(dst, hdr[:]...), text...)
	body := dst[start+dictEntryHdr:]
	maskStream(body, salt, 0)
	crc := crc32.NewIEEE()
	crc.Write(hdr[4:21])
	crc.Write(body)
	binary.LittleEndian.PutUint32(dst[start+21:], crc.Sum32())
	return dst
}

// parseDictEntry returns the intact entry at buf[off:], if there is one, with its text unmasked. An intact entry of a
// kind this version does not know is errUnknownFormat.
func parseDictEntry(buf []byte, off int) (dictEntry, bool, error) {
	b := buf[off:]
	if len(b) < dictEntryHdr || binary.LittleEndian.Uint32(b) != dictMarker {
		return dictEntry{}, false, nil
	}
	kind := b[4]
	id := binary.LittleEndian.Uint32(b[5:])
	n := binary.LittleEndian.Uint32(b[9:])
	size := dictEntryHdr
	if kind == dictText {
		if uint64(n) > uint64(len(b)-dictEntryHdr) {
			return dictEntry{}, false, nil
		}
		size += int(n)
	}
	crc := crc32.NewIEEE()
	crc.Write(b[4:21])
	crc.Write(b[dictEntryHdr:size])
	if crc.Sum32() != binary.LittleEndian.Uint32(b[21:]) {
		return dictEntry{}, false, nil
	}
	e := dictEntry{kind: kind, id: id, count: n, size: size}
	switch kind {
	case dictText:
		if id == math.MaxUint32 {
			return dictEntry{}, false, nil
		}
		text := slices.Clone(b[dictEntryHdr:size])
		maskStream(text, binary.LittleEndian.Uint64(b[13:]), 0)
		e.text, e.count = string(text), 1
	case dictLost, dictReserved:
		if n == 0 || uint64(id)+uint64(n) > math.MaxUint32 {
			return dictEntry{}, false, nil
		}
	default:
		return dictEntry{}, false, fmt.Errorf("%w: entry of kind %d in an enum dict at %d", errUnknownFormat, kind, off)
	}
	return e, true, nil
}

// tornDictEntry reports whether b, the rest of a dictionary behind its last intact entry, is what a crash leaves of
// an entry it appended: the beginning of one, whose length reaches over the end of the file.
func tornDictEntry(b []byte) bool {
	if len(b) < 4 {
		return bytes.HasPrefix(dictMarkerBytes, b)
	}
	if binary.LittleEndian.Uint32(b) != dictMarker {
		return false
	}
	if len(b) < dictEntryHdr {
		return true
	}
	return b[4] == dictText && uint64(binary.LittleEndian.Uint32(b[9:])) > uint64(len(b)-dictEntryHdr)
}

// dictEntrySize returns the size of the entry at buf[off:] by its marker, kind and length, without its CRC, or 0.
func dictEntrySize(buf []byte, off int) int {
	b := buf[off:]
	if len(b) < dictEntryHdr || binary.LittleEndian.Uint32(b) != dictMarker {
		return 0
	}
	if b[4] != dictText {
		return dictEntryHdr
	}
	if n := uint64(binary.LittleEndian.Uint32(b[9:])); n <= uint64(len(b)-dictEntryHdr) {
		return dictEntryHdr + int(n)
	}
	return 0
}

// nextDictEntry returns the offset of the first entry from off on, which has its marker and its CRC, or -1.
func nextDictEntry(buf []byte, off int) int {
	for off < len(buf) {
		i := bytes.Index(buf[off:], dictMarkerBytes)
		if i < 0 {
			return -1
		}
		off += i
		if _, ok, err := parseDictEntry(buf, off); ok || err != nil {
			return off
		}
		off++
	}
	return -1
}

// parse reads a dictionary in the format of this version. It reports a damaged header and every damaged region,
// from the first byte that is no intact entry up to the next intact one, at its beginning, and each lost entry; a rest
// torn by a crash it does not report.
func (d *enumDict) parse(buf []byte) error {
	n := len(buf)
	if n < dictHeaderLen {
		d.cutAt = 0 // a header torn by a crash
		return nil
	}
	intact := string(buf[0:4]) == dictMagic && string(buf[4:8]) == dictFormat &&
		crc32.ChecksumIEEE(buf[:9]) == binary.LittleEndian.Uint32(buf[9:])
	if intact && buf[8] > dictVersion {
		return fmt.Errorf("%w: enum dict %s has version %d", errUnknownFormat, d.path, buf[8])
	}
	if !intact || buf[8] != dictVersion {
		d.report(d.path, 0, reportDictEntry)
	}
	d.header = true
	off := dictHeaderLen
	next := int64(0)    // the smallest id an entry behind may have
	region := int64(-1) // where a damaged region begins, which the next intact entry ends
	for off < n {
		e, ok, err := parseDictEntry(buf, off)
		if err != nil {
			return err
		}
		if ok && int64(e.id) < next {
			ok = false // ids only grow: not one of ours
		}
		if !ok {
			if region < 0 {
				region = int64(off)
			}
			if skip := dictEntrySize(buf, off); skip > 0 {
				if _, ok, _ := parseDictEntry(buf, off+skip); ok {
					off += skip // only the text or the CRC is damaged: the entry behind it follows at its length
					continue
				}
			}
			if off = nextDictEntry(buf, off+1); off < 0 {
				break // nothing intact behind: the tail below
			}
			continue
		}
		at := int64(off)
		if region >= 0 {
			d.report(d.path, region, reportDictEntry)
			at = region
		}
		if int64(e.id) > next {
			d.places = append(d.places, dictPlace{lo: uint32(next), hi: e.id, off: at})
		}
		switch e.kind {
		case dictText:
			d.add(e.id, e.text)
		case dictLost:
			d.report(d.path, int64(off), reportDictEntry)
			d.places = append(d.places, dictPlace{lo: e.id, hi: e.id + e.count, off: int64(off)})
		case dictReserved:
			d.places = append(d.places, dictPlace{lo: e.id, hi: e.id + e.count, off: int64(off)})
		}
		next = int64(e.id) + int64(e.count)
		region = -1
		off += e.size
	}
	d.endOff, d.tail = int64(n), int64(n)
	if region >= 0 {
		// the rest behind the last intact entry: torn by a crash, or damaged; the ids, which its entries may have had,
		// stay unused
		rest := buf[region:]
		reserve := int64(1)
		if tornDictEntry(rest) {
			d.cutAt, d.endOff = region, region
		} else {
			d.report(d.path, region, reportDictEntry)
			reserve = (int64(len(rest)) + dictEntryHdr - 1) / dictEntryHdr
		}
		d.tail = region
		if next+reserve > math.MaxUint32 {
			return fmt.Errorf("tsdb: enum dict %s has no ids left", d.path)
		}
		d.places = append(d.places, dictPlace{lo: uint32(next), hi: uint32(next + reserve), off: region})
		next += reserve
	}
	d.nextID = uint32(next)
	return nil
}

// convert reads a dictionary in the format before N23 and rewrites it in the format of this version, with the same
// ids, before the column takes a value (R-DEC-TS-NAGO-DICT). It is the only reader of that format. An entry, whose
// length fits and whose CRC does not, is damaged and becomes a lost entry, if behind it an intact entry or the end of
// the file follows; else everything from it on is damaged and its ids, as many as entries can lie there, become one
// lost entry. A rest, which ends inside an entry, is what a crash tore, or a cut; its ids become a reserved entry.
// Lost entries are reported at their offsets in the old file, here and in this life, a rest is not. What the old
// format reads as an empty text, also a zeroed region, stays an empty text. Nothing is healed.
//
// The rewrite writes enum.dict.new, syncs it with the directory, renames it to enum.dict and syncs the directory, so
// that a crash leaves the old or the new dictionary complete; the open of a column removes an enum.dict.new left by a
// crash. If one of these steps fails, the load fails, and the old file stays as it is.
func (d *enumDict) convert(buf []byte) error {
	n := len(buf)
	oldEntry := func(off int) (string, int, bool) {
		if n-off < dictOldEntryHdr {
			return "", 0, false
		}
		l := uint64(binary.LittleEndian.Uint32(buf[off:]))
		start := off + dictOldEntryHdr
		if uint64(start)+l > uint64(n) {
			return "", 0, false
		}
		end := start + int(l)
		if crc32.ChecksumIEEE(buf[start:end]) != binary.LittleEndian.Uint32(buf[off+4:]) {
			return "", end, false
		}
		return string(buf[start:end]), end, true
	}
	type rangeOf struct {
		lo, hi uint32
		off    int64
		lost   bool
	}
	var ranges []rangeOf
	id := int64(0)
	off := 0
	tail := int64(n)
	for off < n {
		text, end, ok := oldEntry(off)
		if ok {
			d.add(uint32(id), text)
			id++
			off = end
			continue
		}
		count := int64(1)
		lost := end > 0
		if lost {
			if _, _, behind := oldEntry(end); end != n && !behind {
				end = 0 // the length is not to be trusted: the rest is damaged
			}
		}
		if !lost || end == 0 {
			count = int64(n-off+dictOldEntryHdr-1) / dictOldEntryHdr
		}
		if id+count > math.MaxUint32 {
			return fmt.Errorf("tsdb: enum dict %s has no ids left", d.path)
		}
		if lost {
			d.report(d.path, int64(off), reportDictEntry)
		}
		ranges = append(ranges, rangeOf{lo: uint32(id), hi: uint32(id + count), off: int64(off), lost: lost})
		id += count
		if end == 0 {
			tail = int64(off)
			break
		}
		off = end
	}

	out := dictHeader()
	ri := 0
	for i := uint32(0); int64(i) < id; {
		if ri < len(ranges) && ranges[ri].lo == i {
			r := ranges[ri]
			kind := dictReserved
			if r.lost {
				kind = dictLost
			}
			out = appendDictEntry(out, kind, r.lo, r.hi-r.lo, "")
			d.places = append(d.places, dictPlace{lo: r.lo, hi: r.hi, off: r.off})
			i = r.hi
			ri++
			continue
		}
		s, _ := d.text(i)
		out = appendDictEntry(out, dictText, i, 1, s)
		i++
	}
	d.tail = tail

	dir := filepath.Dir(d.path)
	newPath := d.path + newSuffix
	if err := d.fs.WriteFile(newPath, out); err != nil {
		return fmt.Errorf("tsdb: rewrite enum dict: %w", err)
	}
	if err := d.fs.SyncAll(newPath, dir); err != nil {
		_ = d.fs.Remove(newPath)
		return fmt.Errorf("tsdb: rewrite enum dict: %w", err)
	}
	if err := d.fs.Rename(newPath, d.path); err != nil {
		_ = d.fs.Remove(newPath)
		return fmt.Errorf("tsdb: rewrite enum dict: %w", err)
	}
	if err := d.fs.SyncAll(dir); err != nil {
		return fmt.Errorf("tsdb: rewrite enum dict: %w", err)
	}
	d.header, d.nextID = true, uint32(id)
	d.endOff, d.cutAt = int64(len(out)), -1
	d.size, d.securedEnd, d.dirToSync = int64(len(out)), int64(len(out)), false
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
		if d.byID[v-1] == s && d.has(v-1) {
			return v - 1, true
		}
	}
}

func (d *enumDict) has(id uint32) bool {
	w := int(id / 64)
	return w < len(d.known) && d.known[w]&(1<<(id%64)) != 0
}

// text returns the text of id. Caller holds d.mu.
func (d *enumDict) text(id uint32) (string, bool) {
	if !d.has(id) {
		return "", false
	}
	return d.byID[id], true
}

// add gives s the id. Caller holds d.mu for writing.
func (d *enumDict) add(id uint32, s string) {
	if int(id) >= len(d.byID) {
		d.byID = append(d.byID, make([]string, int(id)+1-len(d.byID))...)
	}
	for int(id/64) >= len(d.known) {
		d.known = append(d.known, 0)
	}
	d.byID[id] = s
	d.known[id/64] |= 1 << (id % 64)
	d.texts++
	if 2*d.texts > len(d.index) {
		n := max(8, 2*len(d.index))
		d.index = make([]uint32, n)
		for i, str := range d.byID {
			if d.has(uint32(i)) {
				d.place(str, uint32(i))
			}
		}
		return
	}
	d.place(s, id)
}

func (d *enumDict) place(s string, id uint32) {
	mask := uint64(len(d.index) - 1)
	i := maphash.String(dictSeed, s) & mask
	for d.index[i] != 0 {
		i = (i + 1) & mask
	}
	d.index[i] = id + 1
}

// intern returns the id for s, appending a new entry if unseen. The append is a single pwrite of the entry, after a
// rest torn by a crash is cut off; it is not waited for (secure).
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
	if d.nextID == math.MaxUint32 {
		return 0, fmt.Errorf("tsdb: enum dict %s has no ids left", d.path)
	}
	if d.cutAt >= 0 {
		if err := d.fs.Truncate(d.path, d.cutAt); err != nil {
			return 0, fmt.Errorf("tsdb: append enum dict: %w", err)
		}
		d.endOff, d.securedEnd = d.cutAt, min(d.securedEnd, d.cutAt)
		d.cutAt = -1
	}
	var frame []byte
	at := d.endOff
	if !d.header {
		frame, at = dictHeader(), 0
	}
	id := d.nextID
	frame = appendDictEntry(frame, dictText, id, 1, s)
	if _, err := d.fs.WriteAt(d.path, frame, at); err != nil {
		return 0, fmt.Errorf("tsdb: append enum dict: %w", err)
	}
	d.header = true
	d.endOff = at + int64(len(frame))
	d.size = d.endOff
	d.nextID++
	d.add(id, s)
	return id, nil
}

// secure puts every entry, which the file holds and which is not on the disk yet, there, with the entry of the file in
// the directory, if that may not be on the disk yet (one SyncAll of both). values.write calls it before a value goes
// into a file of the column; entries of the same time go with one wait. Without such entries it does nothing.
func (d *enumDict) secure() error {
	if d == nil {
		return nil
	}
	d.secMu.Lock()
	defer d.secMu.Unlock()
	paths, end := d.stage()
	if paths == nil {
		return nil
	}
	if err := d.fs.SyncAll(paths...); err != nil {
		return fmt.Errorf("tsdb: sync enum dict: %w", err)
	}
	d.secured(end)
	return nil
}

// stage returns what a sync must write so that the entries, which the file holds now, are on the disk: the file and
// perhaps its directory, and where they end; nil if nothing. A compaction puts them into its SyncAll and calls secured
// after it.
func (d *enumDict) stage() ([]string, int64) {
	if d == nil {
		return nil, 0
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.endOff <= d.securedEnd {
		return nil, 0
	}
	if d.dirToSync {
		return []string{d.path, filepath.Dir(d.path)}, d.endOff
	}
	return []string{d.path}, d.endOff
}

// secured records that the file is on the disk up to end, with its entry in the directory.
func (d *enumDict) secured(end int64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if end > d.securedEnd && end <= d.endOff {
		d.securedEnd = end
		d.dirToSync = false
	}
}

// abandon makes each later intern fail, so that a handle of a dropped column can not write to the dictionary of a new
// column of the same name.
func (d *enumDict) abandon() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.gone = true
}

// lookup resolves an id to its string. ok is false for an id without a text; the caller skips its point and reports
// it (missing).
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
	return d.text(id)
}

// missing reports a point, whose id has no text, at the place of that id: once per place until the DB is opened
// again (Column.reportDamaged).
func (d *enumDict) missing(id uint32) {
	d.mu.RLock()
	off := d.missingPlace(id)
	d.mu.RUnlock()
	d.report(d.path, off, reportDictMissing)
}

// missingPlace is the place of an id without a text: the damaged region, lost or reserved entry or gap of the file,
// where it would be, else where the file read ended (tail): the beginning of a damaged or torn rest, or the end of
// the last entry, 0 for a missing file. Caller holds d.mu.
func (d *enumDict) missingPlace(id uint32) int64 {
	i, found := slices.BinarySearchFunc(d.places, id, func(p dictPlace, id uint32) int {
		switch {
		case id < p.lo:
			return 1
		case id >= p.hi:
			return -1
		}
		return 0
	})
	if found {
		return d.places[i].off
	}
	return d.tail
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
