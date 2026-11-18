// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb_test

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"go.wdy.de/nago/pkg/ndb/tsdb"
)

// What the tests of N23 take from the description of the formats in the code (R-DEC-TS-NAGO-DAMAGE, -DICT, -WAL,
// -SPEED: "Die Lage von Kennzeichen, Art und Länge darf ein Test der Beschreibung des Formats im Code entnehmen", "nach
// der Beschreibung des Kopfs im Code", "Namen der Dateien und Schritte nach der Beschreibung im Code"). Everything else
// the tests observe: where a record, an entry or a frame begins and ends from the size of the file before and after the
// step that wrote it, and the layouts of nago 8ae1daaee, which the code of that stand describes (chunk.go,
// head.go:73–86, enumdict.go:23–27). A text lies with bytes.Index only in a dictionary in the format before N23: from N23
// on the texts are masked (L-N23-36), and a test finds an entry of N23 by its unmasked header (n23DictLayout).
//
// The functions of the format of N23 below are filled in from that description once it is in the code; until then a
// test that needs one fails its premise and says which.

// n23FormatMissing fails the premise of a test, which needs a part of a format of N23 that is not taken from the
// description in the code yet.
func n23FormatMissing(t testing.TB, what string) {
	t.Helper()
	t.Fatalf("premise: %s is not taken from the description of the format of N23 in the code yet", what)
}

// ---- the layouts of nago 8ae1daaee ----

// n23OldDictEntryHdr is the header of an entry of enum.dict before N23: | Len 4 | CRC32 4 over the text | (enumdict.go
// of 8ae1daaee).
const n23OldDictEntryHdr = 8

// n23OldDictEntries returns the offsets of the entries of an enum.dict in the format before N23 and their texts; the
// walk ends at the first entry, which is torn or whose CRC does not match.
func n23OldDictEntries(b []byte) (offs []int, texts []string) {
	for at := 0; at+n23OldDictEntryHdr <= len(b); {
		l := int(binary.LittleEndian.Uint32(b[at:]))
		end := at + n23OldDictEntryHdr + l
		if l < 0 || end > len(b) || crc32.ChecksumIEEE(b[at+n23OldDictEntryHdr:end]) != binary.LittleEndian.Uint32(b[at+4:]) {
			break
		}
		offs, texts = append(offs, at), append(texts, string(b[at+n23OldDictEntryHdr:end]))
		at = end
	}
	return offs, texts
}

// n23OldFirstFrame reads the first frame of b by the layout of the WAL of 8ae1daaee (head.go:73–86: | Kind 1 | Len 4 |
// CRC32 4 over the payload | payload |) and reports its kind and whether it is intact: its length within the file and
// its CRC right.
func n23OldFirstFrame(b []byte) (kind byte, end int, intact bool) {
	if len(b) < 9 {
		return 0, 0, false
	}
	l := int(binary.LittleEndian.Uint32(b[1:]))
	end = 9 + l
	if l < 0 || end > len(b) {
		return b[0], 0, false
	}
	return b[0], end, crc32.ChecksumIEEE(b[9:end]) == binary.LittleEndian.Uint32(b[5:])
}

// n23ChunkPrefixOK reports whether b begins with the prefix of a chunk of version 2 with its CRC: "NAGO" "NTSB", the
// version at byte 8 and the CRC32 (IEEE) of the bytes 0 … 16 at byte 17 (chunk.go:51–53 of 8ae1daaee).
func n23ChunkPrefixOK(b []byte) bool {
	return len(b) >= 21 && string(b[0:4]) == "NAGO" && string(b[4:8]) == "NTSB" &&
		crc32.ChecksumIEEE(b[:17]) == binary.LittleEndian.Uint32(b[17:])
}

// n23SetChunkVersion sets the version of the chunk b and renews the CRC of its prefix.
func n23SetChunkVersion(b []byte, v byte) {
	b[8] = v
	binary.LittleEndian.PutUint32(b[17:], crc32.ChecksumIEEE(b[:17]))
}

// ---- the formats of N23, from the description in the code ----
//
// Taken from the comments of nago tsdb at 230e54235 (N23 Umsetzung, Arbeitsstand): chunk.go (header), head.go (WAL,
// convertLocked) and enumdict.go (dictionary, convert):
//
//	chunk of version 3: the header of version 2, blocks at 85
//	WAL mark:   | Kind 1 | Len 4 = 9 | CRC32 4 over the payload | Payload 9: "NAGO" "NTSW" Version |
//	WAL record: | Marker 4 | Kind 1 | Len 4 | CRC32 4 over Kind, Len and Payload | Payload n |
//	dictionary: | "NAGO" 4 | "NTSD" 4 | Version 1 | CRC32 4 over Magic .. Version |, then entries (154e9035e)
//	            | Marker 4 | Kind 1 | ID 4 | Len 4 | Salt 8 | CRC32 4 over Kind .. Salt and the stored text | Text n masked |
//	rewrites:   head.wal.new and enum.dict.new, renamed over head.wal and enum.dict

// n23ChunkV3DataOff is where the blocks of a chunk of version 3 begin, by the description of the header in the code
// (L-N23-24: "die Lage der Blöcke der Version 3 nach der Beschreibung des Kopfs im Code").
const n23ChunkV3DataOff = 85

// n23WALMark is the length of the mark of a WAL of N23, which a WAL of N23 with records begins with.
const n23WALMark = 18

// n23WALRecord describes a record of a WAL of N23 at b[at:end], whose bounds the test observed: where its kind and its
// length lie and where its payload begins.
type n23WALRecord struct {
	kindOff, lenOff, lenSize, payloadOff int // relative to the start of the record
}

// n23WALLayout returns the layout of a record of a WAL of N23.
func n23WALLayout(t testing.TB) n23WALRecord {
	t.Helper()
	return n23WALRecord{kindOff: 4, lenOff: 5, lenSize: 4, payloadOff: 13}
}

// n23WALRenew renews the checksum of the record of a WAL of N23 at b[at:end] after an edit, so that it is intact.
func n23WALRenew(t testing.TB, b []byte, at, end int) {
	t.Helper()
	sum := crc32.ChecksumIEEE(b[at+4 : at+9])
	sum = crc32.Update(sum, crc32.IEEETable, b[at+13:end])
	binary.LittleEndian.PutUint32(b[at+9:], sum)
}

// n23WALRecords returns the bounds of the records of a WAL of N23 behind its mark, by their lengths; ok is false, if
// one is not intact or the walk does not end at the end of the file.
func n23WALRecords(b []byte) (recs [][2]int, ok bool) {
	if len(b) < n23WALMark || string(b[9:17]) != "NAGONTSW" {
		return nil, false
	}
	marker := uint32(0)
	for at := n23WALMark; at < len(b); {
		if at+13 > len(b) {
			return recs, false
		}
		m := binary.LittleEndian.Uint32(b[at:])
		if marker == 0 {
			marker = m
		}
		end := at + 13 + int(binary.LittleEndian.Uint32(b[at+5:]))
		if m != marker || m == 0 || end > len(b) {
			return recs, false
		}
		sum := crc32.Update(crc32.ChecksumIEEE(b[at+4:at+9]), crc32.IEEETable, b[at+13:end])
		if sum != binary.LittleEndian.Uint32(b[at+9:]) {
			return recs, false
		}
		recs = append(recs, [2]int{at, end})
		at = end
	}
	return recs, true
}

// n23WALIsNew reports whether the WAL b is in the format of N23 throughout: its mark and every record behind it.
func n23WALIsNew(t testing.TB, b []byte) bool {
	t.Helper()
	kind, end, intact := n23OldFirstFrame(b)
	if !intact || kind == 0 || kind == 1 || end != n23WALMark {
		return false
	}
	_, ok := n23WALRecords(b)
	return ok
}

// n23WALBumpVersion sets the version of the marker of the WAL b to one after N23 and renews the checksum of the marker.
func n23WALBumpVersion(t testing.TB, b []byte) {
	t.Helper()
	if len(b) < n23WALMark || string(b[9:17]) != "NAGONTSW" {
		t.Fatalf("premise: the WAL does not begin with the mark of N23")
	}
	b[17]++
	binary.LittleEndian.PutUint32(b[5:], crc32.ChecksumIEEE(b[9:n23WALMark]))
}

// n23WALRewriteFiles are the names of the files, which the rewrite of a WAL in the format before N23 writes besides
// head.wal, in the order of its steps.
func n23WALRewriteFiles(t testing.TB) []string {
	t.Helper()
	return []string{"head.wal.new"}
}

// n23DictEntryHdr is the header of an entry of a dictionary of N23 (enumdict.go at 154e9035e): | Marker 4 | Kind 1 |
// ID 4 | Len 4 | Salt 8 | CRC32 4 over Kind .. Salt and the stored text |; the text of a text entry follows it, masked
// with the key stream of the salt (L-N23-36).
const n23DictEntryHdr = 25

// n23DictCRCOff is where the CRC of an entry of N23 lies, relative to its start.
const n23DictCRCOff = 21

// n23DictHdr is the length of the header of a dictionary of N23 (enumdict.go): | "NAGO" 4 | "NTSD" 4 | Version 1 |
// CRC32 4 |.
const n23DictHdr = 13

// n23DictLay is what a test reads of a dictionary of N23 besides the description: the bytes of the marker of an entry
// and the kind of a text entry, from the first entry of a dictionary of one value.
type n23DictLay struct {
	marker   []byte
	textKind byte
}

// n23DictLayout writes one value of one byte into a new enum column and reads its dictionary: by the description it is
// the header and one text entry of n23DictEntryHdr + 1 bytes, whose first bytes are the marker and the kind of a text.
// A dictionary of another length fails the premise: the description in enumdict.go and the code disagree.
func n23DictLayout(t *testing.T) n23DictLay {
	t.Helper()
	dir := t.TempDir()
	db := openDB(t, dir, tsdb.Options{})
	must(t, column(t, db, "lay", "d", tsdb.Schema{Scheme: tsdb.SchemeEnum}).PutString(1, "x"))
	closeDB(t, db)
	b, err := os.ReadFile(filepath.Join(dir, "lay", "d", "enum.dict"))
	must(t, err)
	if !n23DictIsNew(t, b) || len(b) != n23DictHdr+n23DictEntryHdr+1 {
		t.Fatalf("premise: the dictionary of one value of 1 B has %d B (format of N23 %v); by the description of the format in enumdict.go it has %d: description and code disagree",
			len(b), n23DictIsNew(t, b), n23DictHdr+n23DictEntryHdr+1)
	}
	return n23DictLay{marker: slices.Clone(b[n23DictHdr : n23DictHdr+4]), textKind: b[n23DictHdr+4]}
}

// entries are the starts of the entries of the dictionary b of N23 by their ids, walked by their unmasked headers from
// the end of the file header: a text entry is n23DictEntryHdr bytes and its text of Len, every other kind (lost,
// reserved) n23DictEntryHdr bytes. The walk ends at the first place without the marker.
func (l n23DictLay) entries(b []byte) map[uint32]int {
	res := map[uint32]int{}
	for at := n23DictHdr; at+n23DictEntryHdr <= len(b) && bytes.Equal(b[at:at+4], l.marker); {
		id := binary.LittleEndian.Uint32(b[at+5:])
		if _, ok := res[id]; !ok {
			res[id] = at
		}
		end := at + n23DictEntryHdr
		if b[at+4] == l.textKind {
			end += int(binary.LittleEndian.Uint32(b[at+9:]))
		}
		at = end
	}
	return res
}

// n23DictLenField returns where the length of the entry of enum.dict of N23 beginning at at lies and its size in bytes.
func n23DictLenField(t testing.TB, b []byte, at int) (off, size int) {
	t.Helper()
	return at + 9, 4
}

// n23DictIsNew reports whether the dictionary b is in the format of N23.
func n23DictIsNew(t testing.TB, b []byte) bool {
	t.Helper()
	return len(b) >= 13 && string(b[0:8]) == "NAGONTSD"
}

// n23DictBumpVersion sets the version of the dictionary b of N23 to one after N23 and renews the checksum over it.
func n23DictBumpVersion(t testing.TB, b []byte) {
	t.Helper()
	if !n23DictIsNew(t, b) {
		t.Fatalf("premise: the dictionary does not begin with the header of N23")
	}
	b[8]++
	binary.LittleEndian.PutUint32(b[9:], crc32.ChecksumIEEE(b[:9]))
}

// n23DictFormatPlace is a byte of the dictionary b of N23, at which nago tells the format of N23 from the one before it:
// the magic "NAGO" of the header (the format "NTSD" keeps the file in the format of N23).
func n23DictFormatPlace(t testing.TB, b []byte) int {
	t.Helper()
	return 0
}

// n23DictRewriteFiles are the names of the files, which the rewrite of a dictionary in the format before N23 writes
// besides enum.dict, in the order of its steps.
func n23DictRewriteFiles(t testing.TB) []string {
	t.Helper()
	return []string{"enum.dict.new"}
}

// n23EditWALKind sets the kind of the record with the index i of the WAL of N23 at path and renews its checksum, so
// that it stays intact (TestN22UnknownAborts, wokoda N23, L-N23-24).
func n23EditWALKind(t testing.TB, path string, i int, kind byte) n22Unit {
	t.Helper()
	b := n22ReadAt(t, path)
	recs, ok := n23WALRecords(b)
	if !ok || i >= len(recs) {
		t.Fatalf("premise: the WAL holds %d intact records of N23 (all intact %v), the test edits number %d", len(recs), ok, i)
	}
	at, end := recs[i][0], recs[i][1]
	if k := b[at+4]; k != 0 && k != 1 {
		t.Fatalf("premise: the record at %d has the kind %#x, not 0 or 1", at, k)
	}
	b[at+4] = kind
	n23WALRenew(t, b, at, end)
	must(t, os.WriteFile(path, b, 0o644))
	return n22Unitof(path, b, at, end)
}

// n23WithoutDict drops the differences of enum.dict from a list of n22SnapshotDiff: a read, which rewrites a dictionary
// in the format before N23, changes it and no other file of the column (R-DEC-TS-NAGO-DICT, L-N23-24).
func n23WithoutDict(diffs []string) []string {
	var out []string
	for _, d := range diffs {
		if len(d) < 10 || d[:10] != "enum.dict " {
			out = append(out, d)
		}
	}
	return out
}
