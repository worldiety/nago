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
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"

	"go.wdy.de/nago/pkg/ndb"
)

// chunk file header, matching the house style of msgstore:
//
//	| Magic        4 | 0x4E41474F "NAGO"
//	| Format Magic 4 | 0x4E545342 "NTSB" Nago Time Series Block
//	| Version      1 | 0x01 or 0x02
//
// Version 1 has nothing more; its blocks begin at byte 9. Version 2 goes on with
//
//	| Salt    8 | random, the key of the masks of the file (see maskStream)
//	| CRC32   4 | IEEE over Magic .. Salt
//	| Slot A 32 |
//	| Slot B 32 |
//
// and its blocks begin at byte 85. A slot tells how far the chunk was sealed, so that the open of a column reads only
// the end of its open chunk instead of the whole file:
//
//	| Seq     4 | the slot with the larger number is the newer one
//	| End     8 | end of the sealed blocks
//	| Points  8 | points of the sealed blocks
//	| Max     8 | newest time of the sealed blocks
//	| CRC32   4 | IEEE over Seq .. Max
//
// A seal writes the slot, which is not the newer one, so that a torn write leaves the other intact. The bodies of the
// blocks and continuation records of a version 2 file are masked with a key stream of the salt and their offset, so
// that no value, which a caller writes, appears in the file as written: the search for blocks and the recovery, which
// look for markers, can not take the bytes of a value for a frame. Every file tsdb writes has version 2; it reads
// both, and continues a version 1 file unmasked. Every later version keeps the prefix of version 2 up to its CRC, so
// that a chunk of a later version is errUnknownFormat, not a corrupt file (readChunkHeader); every later change of the
// continuation records, the WAL or the dictionary increases the version, a new encoding of a block gets a flag.
const (
	chunkMagic      uint32 = 0x4E41474F // "NAGO"
	chunkFormat     uint32 = 0x4E545342 // "NTSB"
	chunkVersion1   byte   = 1
	chunkVersion    byte   = 2
	chunkPrefixLen         = 9
	chunkV2Prefix          = chunkPrefixLen + 8 + 4
	chunkSlotLen           = 4 + 8 + 8 + 8 + 4
	chunkV1DataOff  int64  = chunkPrefixLen
	chunkDataOff    int64  = chunkV2Prefix + 2*chunkSlotLen
	chunkExt               = ".tsb"
	chunkSearchFrom int64  = 256 << 10 // blocks regions up to this size are read as a whole
)

// chunkHead is what the header of a chunk file tells.
type chunkHead struct {
	dataOff  int64  // where the blocks begin
	salt     uint64 // the key of the masks, version 2 only
	masked   bool   // version 2: bodies are masked
	slot     chunkSlot
	haveSlot bool
}

// chunkInfo describes a chunk file on disk. A finalized chunk is named
// "<minMillis>_<maxMillis>.tsb"; a pending (open) chunk is "<minMillis>_.tsb".
// Filenames encode the contained time range so range queries skip whole files
// and retention drops whole files, without opening them — the timezone-free
// analog of msgstore's seq-range filenames.
type chunkInfo struct {
	path      string
	minMillis int64
	maxMillis int64 // for pending chunks this is the last known max (may grow)
	pending   bool
	// sizeBytes is where the blocks of the chunk end: the size of a finalized chunk, the end of the sealed blocks of
	// the open one; 0 is unknown.
	sizeBytes int64
	// dataOff is where its blocks begin, known from its header; 0 until the header was read. salt and masked are those
	// of the header (chunkHead).
	dataOff int64
	salt    uint64
	masked  bool
	// bad is set when the header is not one of a chunk file; reads skip the chunk then.
	bad bool
}

func pendingChunkName(minMillis int64) string {
	return strconv.FormatInt(minMillis, 10) + "_" + chunkExt
}

func finalChunkName(minMillis, maxMillis int64) string {
	return strconv.FormatInt(minMillis, 10) + "_" + strconv.FormatInt(maxMillis, 10) + chunkExt
}

// parseChunkName parses a chunk filename into its time bounds. ok is false for
// files that are not chunk files.
func parseChunkName(name string) (info chunkInfo, ok bool) {
	if !strings.HasSuffix(name, chunkExt) {
		return chunkInfo{}, false
	}
	base := strings.TrimSuffix(name, chunkExt)
	us := strings.IndexByte(base, '_')
	if us < 0 {
		return chunkInfo{}, false
	}
	minStr, maxStr := base[:us], base[us+1:]
	minV, err := strconv.ParseInt(minStr, 10, 64)
	if err != nil {
		return chunkInfo{}, false
	}
	if maxStr == "" {
		return chunkInfo{minMillis: minV, maxMillis: minV, pending: true}, true
	}
	maxV, err := strconv.ParseInt(maxStr, 10, 64)
	if err != nil {
		return chunkInfo{}, false
	}
	return chunkInfo{minMillis: minV, maxMillis: maxV}, true
}

// chunkSlot is a state slot of a version 2 header.
type chunkSlot struct {
	seq    uint32
	end    int64
	points int64
	max    int64
}

func (s chunkSlot) encode() [chunkSlotLen]byte {
	var b [chunkSlotLen]byte
	binary.LittleEndian.PutUint32(b[0:], s.seq)
	binary.LittleEndian.PutUint64(b[4:], uint64(s.end))
	binary.LittleEndian.PutUint64(b[12:], uint64(s.points))
	binary.LittleEndian.PutUint64(b[20:], uint64(s.max))
	binary.LittleEndian.PutUint32(b[28:], crc32.ChecksumIEEE(b[:28]))
	return b
}

func decodeSlot(b []byte) (chunkSlot, bool) {
	if len(b) < chunkSlotLen || crc32.ChecksumIEEE(b[:28]) != binary.LittleEndian.Uint32(b[28:]) {
		return chunkSlot{}, false
	}
	return chunkSlot{
		seq:    binary.LittleEndian.Uint32(b[0:]),
		end:    int64(binary.LittleEndian.Uint64(b[4:])),
		points: int64(binary.LittleEndian.Uint64(b[12:])),
		max:    int64(binary.LittleEndian.Uint64(b[20:])),
	}, true
}

// slotOffset is the offset of the slot, which a header with the newest slot seq writes next.
func slotOffset(seq uint32) int64 { return chunkV2Prefix + int64(seq%2)*chunkSlotLen }

// chunkHeader returns a version 2 header with salt whose first slot is s.
func chunkHeader(s chunkSlot, salt uint64) []byte {
	hdr := make([]byte, chunkDataOff)
	binary.BigEndian.PutUint32(hdr[0:], chunkMagic)
	binary.BigEndian.PutUint32(hdr[4:], chunkFormat)
	hdr[8] = chunkVersion
	binary.LittleEndian.PutUint64(hdr[chunkPrefixLen:], salt)
	binary.LittleEndian.PutUint32(hdr[chunkPrefixLen+8:], crc32.ChecksumIEEE(hdr[:chunkPrefixLen+8]))
	slot := s.encode()
	copy(hdr[slotOffset(s.seq):], slot[:])
	return hdr
}

// newSalt returns the salt of a new chunk file.
func newSalt() uint64 { return rand.Uint64() }

// writeSlot writes s into its place of the header of path.
func writeSlot(fs *ndb.FilePool, path string, s chunkSlot) error {
	b := s.encode()
	if _, err := fs.WriteAt(path, b[:], slotOffset(s.seq)); err != nil {
		return fmt.Errorf("tsdb: write chunk slot: %w", err)
	}
	return nil
}

// readChunkHeader reads the header of path: where the blocks begin, the salt and, for version 2, the newer intact
// slot. errCorruptChunk reports a file that is not a chunk, errUnknownFormat one of a later version: its magic and its
// format match, its version is larger than 2, and the CRC of the prefix of version 2 (magic .. salt) matches or the
// file is shorter than a header of version 2, so that the open of a column does not remove it as a torn header. Every
// later version keeps that prefix.
func readChunkHeader(fs *ndb.FilePool, path string) (chunkHead, error) {
	var hdr [chunkDataOff]byte
	n, rerr := fs.ReadAt(path, hdr[:], 0)
	if rerr != nil && rerr != io.EOF {
		return chunkHead{}, fmt.Errorf("tsdb: read chunk header: %w", rerr)
	}
	if n < chunkPrefixLen || binary.BigEndian.Uint32(hdr[0:]) != chunkMagic || binary.BigEndian.Uint32(hdr[4:]) != chunkFormat {
		return chunkHead{}, errCorruptChunk
	}
	switch hdr[8] {
	case chunkVersion1:
		return chunkHead{dataOff: chunkV1DataOff}, nil
	case chunkVersion:
		if n < int(chunkDataOff) || crc32.ChecksumIEEE(hdr[:chunkPrefixLen+8]) != binary.LittleEndian.Uint32(hdr[chunkPrefixLen+8:]) {
			return chunkHead{}, errCorruptChunk
		}
		h := chunkHead{dataOff: chunkDataOff, salt: binary.LittleEndian.Uint64(hdr[chunkPrefixLen:]), masked: true}
		a, aok := decodeSlot(hdr[chunkV2Prefix:])
		b, bok := decodeSlot(hdr[chunkV2Prefix+chunkSlotLen:])
		switch {
		case aok && bok:
			if b.seq > a.seq {
				a = b
			}
			h.slot, h.haveSlot = a, true
		case aok:
			h.slot, h.haveSlot = a, true
		case bok:
			h.slot, h.haveSlot = b, true
		}
		return h, nil
	}
	if hdr[8] > chunkVersion && (n < int(chunkDataOff) ||
		crc32.ChecksumIEEE(hdr[:chunkPrefixLen+8]) == binary.LittleEndian.Uint32(hdr[chunkPrefixLen+8:])) {
		// a later version, which keeps the prefix of version 2 with its CRC, or a file of a later version, which is
		// shorter than a header of version 2 and is not taken for a torn header
		return chunkHead{}, fmt.Errorf("%w: chunk %s has version %d", errUnknownFormat, path, hdr[8])
	}
	return chunkHead{}, errCorruptChunk
}

// maskStream XORs b with a key stream of salt and off (splitmix64), the offset of the frame of b in the file. It is
// its own inverse. The stream is not secret against someone who reads the file; it keeps a writer of values, who does
// not know the salt, from placing chosen bytes into the file.
func maskStream(b []byte, salt uint64, off int64) {
	golden := uint64(0x9E3779B97F4A7C15)
	g2, g3, g4 := golden*2, golden*3, golden*4 // wrapping like the additions of splitmix64
	x := salt ^ uint64(off)*golden
	// four words at a time: the words of the stream do not depend on each other, only on their number
	for len(b) >= 32 {
		x0, x1, x2, x3 := x+golden, x+g2, x+g3, x+g4
		x += g4
		w := b[:32]
		binary.LittleEndian.PutUint64(w[0:], binary.LittleEndian.Uint64(w[0:])^splitmix(x0))
		binary.LittleEndian.PutUint64(w[8:], binary.LittleEndian.Uint64(w[8:])^splitmix(x1))
		binary.LittleEndian.PutUint64(w[16:], binary.LittleEndian.Uint64(w[16:])^splitmix(x2))
		binary.LittleEndian.PutUint64(w[24:], binary.LittleEndian.Uint64(w[24:])^splitmix(x3))
		b = b[32:]
	}
	for len(b) >= 8 {
		x += golden
		binary.LittleEndian.PutUint64(b, binary.LittleEndian.Uint64(b)^splitmix(x))
		b = b[8:]
	}
	if len(b) > 0 {
		x += golden
		z := splitmix(x)
		for i := range b {
			b[i] ^= byte(z >> (8 * i))
		}
	}
}

// splitmix is the output function of splitmix64 for the state x.
func splitmix(x uint64) uint64 {
	z := (x ^ (x >> 30)) * 0xBF58476D1CE4E5B9
	z = (z ^ (z >> 27)) * 0x94D049BB133111EB
	return z ^ (z >> 31)
}

// maskBlock masks the body of the encoded block frame, which goes to offset at, sets flagMasked and renews its CRC,
// which covers the masked bytes.
func maskBlock(frame []byte, salt uint64, at int64) {
	region := frame[blkHeaderSz : len(frame)-blkCRC]
	region[blkPreBodySz-1] |= flagMasked
	maskStream(region[blkPreBodySz:], salt, at)
	binary.LittleEndian.PutUint32(frame[len(frame)-blkCRC:], crc32.ChecksumIEEE(region))
}

// unmaskBlock undoes maskBlock on an intact frame in place, after which its CRC no longer matches; decode it with
// decodeBlockData.
func unmaskBlock(frame []byte, h blockHeader, salt uint64, at int64) {
	if h.flags&flagMasked != 0 {
		maskStream(frame[h.bodyOff:h.bodyOff+h.bodyLen], salt, at)
	}
}

// ---- continuation records ----

// recordMarker begins a continuation record. A Sync appends one behind the sealed blocks of the open chunk with the
// points written since the Sync before; a seal writes one, which copies all saved points of the open block, before it
// overwrites the records with the block. Their marker differs from the one of blocks, so that a block scan never takes
// a record for a block.
//
//	| Marker   8 | recordMarker
//	| BodyLen  4 | length of Kind .. Body
//	| Kind     1 | recordDelta or recordCopy
//	| Count    4 | number of points
//	| Body     n | the points (see pointEnc)
//	| CRC32    4 | IEEE over Kind .. Body
const recordMarker uint64 = 0xDEAD54534443BEEF // "…TSDC…"

const (
	recordDelta  byte = 0    // the points since the record before
	recordCopy   byte = 1    // all saved points of the open block, it replaces the records before it
	recordMasked byte = 0x80 // the body is masked (version 2 files)
	recHeadLen        = 8 + 4
	recPreLen         = 1 + 4
	recCRCLen         = 4
	recOverhead       = recHeadLen + recPreLen + recCRCLen
)

// appendRecord appends the record frame of the encoded points body to dst; with masked its body is masked for the
// offset at, where the frame goes.
func appendRecord(dst []byte, kind byte, count int, body []byte, masked bool, salt uint64, at int64) []byte {
	var hdr [recHeadLen + recPreLen]byte
	binary.LittleEndian.PutUint64(hdr[0:], recordMarker)
	binary.LittleEndian.PutUint32(hdr[8:], uint32(recPreLen+len(body)))
	if masked {
		kind |= recordMasked
	}
	hdr[12] = kind
	binary.LittleEndian.PutUint32(hdr[13:], uint32(count))
	start := len(dst)
	dst = append(dst, hdr[:]...)
	dst = append(dst, body...)
	if masked {
		maskStream(dst[start+recHeadLen+recPreLen:], salt, at)
	}
	crc := crc32.ChecksumIEEE(dst[start+recHeadLen:])
	return binary.LittleEndian.AppendUint32(dst, crc)
}

// frame kinds found by frameReader
const (
	frameNone = iota
	frameBlock
	frameRecord
)

// peekFrame reads the marker and the length of the frame at buf[0]: its kind and total length, or frameNone.
func peekFrame(buf []byte) (kind int, frameLen int) {
	if len(buf) < recHeadLen {
		return frameNone, 0
	}
	bodyLen := int(binary.LittleEndian.Uint32(buf[8:]))
	switch binary.LittleEndian.Uint64(buf) {
	case syncMarker:
		if bodyLen < blkPreBodySz {
			return frameNone, 0
		}
		return frameBlock, blkHeaderSz + bodyLen + blkCRC
	case recordMarker:
		if bodyLen < recPreLen {
			return frameNone, 0
		}
		return frameRecord, recHeadLen + bodyLen + recCRCLen
	}
	return frameNone, 0
}

// parseRecord checks the record frame of length frameLen at buf[0], which lies at offset at of a file with salt, and
// returns its kind, count and body, unmasked in place. See unknownRecord for an intact record of a kind of a later
// version.
func parseRecord(buf []byte, frameLen int, salt uint64, at int64) (kind byte, count int, body []byte, ok bool) {
	if len(buf) < frameLen || frameLen < recOverhead {
		return 0, 0, nil, false
	}
	region := buf[recHeadLen : frameLen-recCRCLen]
	if crc32.ChecksumIEEE(region) != binary.LittleEndian.Uint32(buf[frameLen-recCRCLen:frameLen]) {
		return 0, 0, nil, false
	}
	kind = region[0] &^ recordMasked
	if kind != recordDelta && kind != recordCopy {
		return kind, 0, nil, false
	}
	body = region[recPreLen:]
	if region[0]&recordMasked != 0 {
		maskStream(body, salt, at)
		region[0] &^= recordMasked // unmasked now; a second parse would fail its CRC anyway
	}
	return kind, int(binary.LittleEndian.Uint32(region[1:])), body, true
}

// unknownRecord reports, whether the record frame at buf[0] has its CRC and a kind, which this version does not know:
// a later version wrote it. The CRC covers the kind.
func unknownRecord(buf []byte, frameLen int) bool {
	if !recordIntact(buf, frameLen) {
		return false
	}
	kind := buf[recHeadLen] &^ recordMasked
	return kind != recordDelta && kind != recordCopy
}

// recordIntact reports whether the record frame at buf[0] has its CRC.
func recordIntact(buf []byte, frameLen int) bool {
	if len(buf) < frameLen || frameLen < recOverhead {
		return false
	}
	return crc32.ChecksumIEEE(buf[recHeadLen:frameLen-recCRCLen]) == binary.LittleEndian.Uint32(buf[frameLen-recCRCLen:frameLen])
}

// ---- reading frames ----

const (
	readerFirstBatch = 64 << 10
	readerMaxBatch   = 256 << 10
	probeWindow      = 4 << 10
	searchStop       = 64 << 10 // the binary search ends when less than this is left
)

var (
	blockMarkerBytes  = binary.LittleEndian.AppendUint64(nil, syncMarker)
	recordMarkerBytes = binary.LittleEndian.AppendUint64(nil, recordMarker)
)

// frameReader reads the frames of a file between two offsets through the pool, a window at a time, so that a scan
// holds at most a window and a frame in memory and reads only the part of the file it needs.
type frameReader struct {
	col  *Column // whose damaged frames r reports
	fs   *ndb.FilePool
	path string
	end  int64 // where the frames end
	// torn: the frames are those of an open chunk after an earlier life (recoverOpenChunk), where a frame may run over
	// the end, torn by a crash, and bytes between frames may be the rest of records written over; elsewhere the frames
	// follow each other up to the end
	torn bool

	buf   []byte // the bytes of the file from base on
	base  int64
	pos   int64 // where the next frame begins
	batch int
}

// reset makes r read the frames of the file path of c from pos to end; torn as at frameReader.torn.
func (r *frameReader) reset(c *Column, path string, pos, end int64, torn bool) {
	r.col, r.fs, r.path, r.end, r.torn = c, c.pool, path, end, torn
	r.buf, r.base, r.pos, r.batch = r.buf[:0], pos, pos, readerFirstBatch
}

// fill makes at least n bytes from pos on available, as far as the end allows, reading at least want bytes, and
// returns the bytes from pos on.
func (r *frameReader) fill(n, want int) ([]byte, error) {
	if r.pos < r.base || r.pos > r.base+int64(len(r.buf)) {
		r.buf, r.base = r.buf[:0], r.pos
	}
	have := int(r.base + int64(len(r.buf)) - r.pos)
	if have >= n || r.pos+int64(have) >= r.end {
		return r.buf[r.pos-r.base:], nil
	}
	// keep what is left from pos on and read behind it
	copy(r.buf, r.buf[r.pos-r.base:])
	r.buf, r.base = r.buf[:have], r.pos
	read := max(n-have, want)
	if rest := r.end - (r.pos + int64(have)); int64(read) > rest {
		read = int(rest)
	}
	if cap(r.buf) < have+read {
		nb := make([]byte, have, have+read)
		copy(nb, r.buf)
		r.buf = nb
	}
	m, err := r.fs.ReadAt(r.path, r.buf[have:have+read], r.pos+int64(have))
	r.buf = r.buf[:have+m]
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("tsdb: read chunk: %w", err)
	}
	if m < read {
		r.end = r.pos + int64(have+m) // the file is shorter than expected
	}
	return r.buf[r.pos-r.base:], nil
}

// nextBatch is the size of the next sequential read, growing up to readerMaxBatch.
func (r *frameReader) nextBatch() int {
	b := r.batch
	r.batch = min(r.batch*2, readerMaxBatch)
	return b
}

// next returns the next intact frame from pos on, of the kinds in want, skipping everything that is not one, and
// moves pos behind it. kind is frameNone at the end. The returned frame is valid until the next call. Where it skips a
// damaged frame, it reports the place (Column.reportDamaged): a frame of a wanted kind, which fails its CRC, and
// without torn also one, which runs over the end, and bytes, which are no frame at all (with torn, recoverOpenChunk
// reports bytes between two records). A frame of a kind not wanted is skipped without a report.
func (r *frameReader) next(want int) (kind int, at int64, frame []byte, err error) {
	for r.pos < r.end {
		win, err := r.fill(recHeadLen, r.nextBatch())
		if err != nil {
			return frameNone, 0, nil, err
		}
		if len(win) < recHeadLen {
			return frameNone, 0, nil, nil
		}
		k, frameLen := peekFrame(win)
		damaged := !r.torn && (k == frameNone || k&want != 0) // no frame, or one over the end
		if k != frameNone && k&want != 0 && r.pos+int64(frameLen) <= r.end {
			if win, err = r.fill(frameLen, r.nextBatch()); err != nil {
				return frameNone, 0, nil, err
			}
			if len(win) >= frameLen {
				if frameIntact(k, win[:frameLen]) {
					at = r.pos
					r.pos += int64(frameLen)
					return k, at, win[:frameLen], nil
				}
				damaged = true // a wrong CRC
			}
		}
		if damaged {
			r.col.reportDamaged(r.path, r.pos)
		}
		if err := r.skip(want); err != nil {
			return frameNone, 0, nil, err
		}
	}
	return frameNone, 0, nil, nil
}

// skip moves pos to the next marker of a wanted kind behind pos, or to the end.
func (r *frameReader) skip(want int) error {
	r.pos++
	for r.pos < r.end {
		win, err := r.fill(len(blockMarkerBytes), r.nextBatch())
		if err != nil {
			return err
		}
		if len(win) < len(blockMarkerBytes) {
			r.pos = r.end
			return nil
		}
		i := indexMarker(win, want)
		if i >= 0 {
			r.pos += int64(i)
			return nil
		}
		r.pos += int64(len(win) - len(blockMarkerBytes) + 1)
	}
	return nil
}

// indexMarker returns the index of the first marker of a wanted kind in win, or -1.
func indexMarker(win []byte, want int) int {
	i := -1
	if want&frameBlock != 0 {
		i = bytes.Index(win, blockMarkerBytes)
	}
	if want&frameRecord != 0 {
		if j := bytes.Index(win, recordMarkerBytes); j >= 0 && (i < 0 || j < i) {
			i = j
		}
	}
	return i
}

func frameIntact(kind int, frame []byte) bool {
	switch kind {
	case frameBlock:
		h, err := parseBlockHeader(frame)
		return err == nil && blockIntact(frame, h)
	case frameRecord:
		return recordIntact(frame, len(frame))
	}
	return false
}

// search returns a block boundary in [lo, hi], at or before the first block whose newest time is not before qmin, by
// a binary search over the file offsets: each probe reads a small window, so that a query in a large chunk reads a few
// windows instead of every block before its range. lo must be a block boundary. A block taken as a waypoint lies in
// [cmin, cmax], the bounds of the chunk, and has its CRC; since the bodies are masked (maskBlock), no value written
// into the chunk can pose as one.
func (r *frameReader) search(lo, hi, qmin, cmin, cmax int64) (int64, error) {
	for hi-lo > searchStop {
		mid := lo + (hi-lo)/2
		h, at, ok, err := r.probe(mid, hi, cmin, cmax)
		if err != nil {
			return 0, err
		}
		switch {
		case !ok:
			hi = mid
		case h.maxMillis < qmin:
			lo = at + int64(h.frameLen) // every block up to this one ends before qmin
		default:
			hi = mid
		}
	}
	return lo, nil
}

// probe finds the first block starting in [from, to): its marker, a length within the chunk, bounds within [cmin,
// cmax] and an intact CRC, whatever its size.
func (r *frameReader) probe(from, to, cmin, cmax int64) (blockHeader, int64, bool, error) {
	r.pos = from
	for r.pos < to {
		win, err := r.fill(len(blockMarkerBytes), probeWindow)
		if err != nil {
			return blockHeader{}, 0, false, err
		}
		if len(win) < len(blockMarkerBytes) {
			return blockHeader{}, 0, false, nil
		}
		i := bytes.Index(win, blockMarkerBytes)
		if i < 0 {
			r.pos += int64(len(win) - len(blockMarkerBytes) + 1)
			continue
		}
		r.pos += int64(i)
		if r.pos >= to {
			return blockHeader{}, 0, false, nil
		}
		if win, err = r.fill(blkHeaderSz+blkPreBodySz, probeWindow); err != nil {
			return blockHeader{}, 0, false, err
		}
		if k, frameLen := peekFrame(win); k == frameBlock && r.pos+int64(frameLen) <= r.end && len(win) >= blkHeaderSz+blkPreBodySz {
			region := win[blkHeaderSz:]
			minT := int64(binary.LittleEndian.Uint64(region))
			maxT := int64(binary.LittleEndian.Uint64(region[8:]))
			count := int(binary.LittleEndian.Uint32(region[16:]))
			if count > 0 && cmin <= minT && minT <= maxT && maxT <= cmax {
				if win, err = r.fill(frameLen, frameLen); err != nil {
					return blockHeader{}, 0, false, err
				}
				if len(win) >= frameLen && frameIntact(frameBlock, win[:frameLen]) {
					h, _ := parseBlockHeader(win[:frameLen])
					return h, r.pos, true, nil
				}
			}
		}
		r.pos++
	}
	return blockHeader{}, 0, false, nil
}

// statSize returns the size of the file at path.
func statSize(fs *ndb.FilePool, path string) (int64, error) {
	fi, err := fs.Stat(path)
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}

// isNotExist reports a missing file also behind the wrapping of the pool.
func isNotExist(err error) bool { return errors.Is(err, os.ErrNotExist) }
