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
	"sync"

	"github.com/klauspost/compress/s2"
	"go.wdy.de/nago/pkg/xbytes"
)

// syncMarker is the 8-byte resync anchor at the start of every block frame. It
// lets a reader forward-scan past corruption (bitrot / partial writes) to the
// next valid block, matching the recovery approach of msgstore.
const syncMarker uint64 = 0xDEAD54534442BEEF // "…TSDB…"

// block frame layout (all fixed fields little-endian to match xbytes.Buffer):
//
//	| SyncMarker  8 | resync anchor
//	| BodyLen     4 | length of the block body (everything after this field, excl. CRC)
//	| MinMillis   8 | smallest timestamp in the block (range-skip without decoding)
//	| MaxMillis   8 | largest timestamp in the block
//	| Count       4 | number of points
//	| Flags       1 | see flagCompressed .. flagFORVals
//	| Body        n | timestamps stream + values stream (optionally compressed)
//	| CRC32       4 | IEEE over [MinMillis .. Body]
const (
	blkSync     = 8
	blkBodyLen  = 4
	blkMin      = 8
	blkMax      = 8
	blkCount    = 4
	blkFlags    = 1
	blkCRC      = 4
	blkHeaderSz = blkSync + blkBodyLen // bytes before the CRC-covered region
	// CRC covers MinMillis, MaxMillis, Count, Flags and Body.
	blkPreBodySz = blkMin + blkMax + blkCount + blkFlags
)

const (
	// flagCompressed marks the block body as s2-compressed.
	flagCompressed byte = 1 << 0
	// flagEquidistantTS marks the timestamp stream as an equidistant run encoded
	// as just (start, step): ts[i] = start + i*step. Set only when every
	// consecutive delta is identical (no holes within the block). Decodes by
	// closed-form formula with no varint per point.
	flagEquidistantTS byte = 1 << 1
	// flagConstVal marks the numeric value stream (decimal values or enum ids)
	// as a single constant repeated for every point. The stream stores just one
	// varint and the reader fills the whole slice with it. Set only when every
	// value in the block is identical. Not used for string columns.
	flagConstVal byte = 1 << 2
	// flagMasked marks a body masked with the key stream of its file (see maskBlock); readers unmask it before they
	// decode it.
	flagMasked byte = 1 << 3
	// flagFORVals marks the numeric value stream as the frame-of-reference stream of forcodec.go instead of the
	// delta + zig-zag varint stream. A block holds it, if it is not larger than the varint stream (see encodeBlock).
	flagFORVals byte = 1 << 4
	// flagsKnown are the flags of this version. A block, which is intact and has another flag, was written by a later
	// version: a read reports errUnknownFormat instead of skipping it.
	flagsKnown = flagCompressed | flagEquidistantTS | flagConstVal | flagMasked | flagFORVals
)

// blockData is the decoded, in-memory content of a block. For numeric schemes
// (decimal and enum) vals holds the scaled int64 values or enum dictionary ids;
// strs is nil. For the string scheme strs holds one string per timestamp and
// vals is nil.
type blockData struct {
	ts   []int64
	vals []int64  // numeric schemes: decimal values or enum ids
	strs []string // string scheme: raw strings

	decompBuf []byte // reusable s2-decompression target for compressed blocks
}

func (b *blockData) count() int { return len(b.ts) }

// encodeScratch holds the buffers of encodeBlock, so that sealing a block allocates nothing per block.
type encodeScratch struct {
	body, alt   []byte // the varint body and the body with the frame-of-reference values
	zBody, zAlt []byte // their s2-compressed forms
	fs          forScratch
}

var encodePool = sync.Pool{New: func() any { return new(encodeScratch) }}

const maxPooledEncode = 1 << 20 // a larger encode buffer is not kept

// encodeBlock serializes bd into a self-describing block frame appended to out
// and returns the extended slice. compress controls s2 compression of the body.
//
// The values of a numeric block, which are not all the same, are encoded in two ways, as delta + zig-zag varints and
// as the frame-of-reference stream (see forcodec.go), each s2-compressed if that makes it smaller, and the block takes
// the smaller one, the frame-of-reference stream on a tie. So no block is larger than with varints alone.
func encodeBlock(out []byte, bd *blockData, s Scheme, compress bool) ([]byte, error) {
	if !s.valid() {
		return out, errInvalidScheme
	}
	es := encodePool.Get().(*encodeScratch)
	defer putEncodeScratch(es)

	// choose the timestamp codec: a block whose timestamps have a single
	// constant step (equidistant, no holes) is stored as just (start, step),
	// which the reader reconstructs by ts[i] = start + i*step. This drops the
	// per-point timestamp byte entirely and makes decode a pure arithmetic loop.
	equi, step := isEquidistant(bd.ts)

	// choose the value codec: a block whose numeric values are all identical
	// (a signal parked at a setpoint, an idle status, ...) is stored as a single
	// constant and filled on read, collapsing the whole value stream to one
	// varint. Enum ids are just int64 values on the same path as decimal; only
	// strings are handled separately.
	numeric := s != SchemeString
	constVal := numeric && allEqualI64(bd.vals)

	body := es.body[:0]
	if equi {
		body = appendVarint(body, bd.ts[0])
		body = appendVarint(body, step)
	} else {
		body = appendTimestamps(body, bd.ts)
	}
	tsLen := len(body)
	switch {
	case !numeric:
		for _, str := range bd.strs {
			body = binary.AppendUvarint(body, uint64(len(str)))
			body = append(body, str...)
		}
	case constVal:
		body = appendVarint(body, firstI64(bd.vals))
	default:
		body = appendValueDeltas(body, bd.vals)
	}
	es.body = body

	flags := byte(0)
	if equi {
		flags |= flagEquidistantTS
	}
	if constVal {
		flags |= flagConstVal
	}
	payload, pflags := body, flags
	if compress && len(body) > 512 {
		es.zBody = s2.Encode(es.zBody[:cap(es.zBody)], body)
		if len(es.zBody) < len(body) {
			payload, pflags = es.zBody, flags|flagCompressed
		}
	}
	if numeric && !constVal {
		alt := append(es.alt[:0], body[:tsLen]...)
		alt = appendFORValues(alt, bd.vals, &es.fs)
		es.alt = alt
		altPayload, altFlags := alt, flags|flagFORVals
		if len(alt) > len(payload) && compress && len(alt) > 512 {
			es.zAlt = s2.Encode(es.zAlt[:cap(es.zAlt)], alt)
			if len(es.zAlt) < len(alt) {
				altPayload, altFlags = es.zAlt, altFlags|flagCompressed
			}
		}
		if len(altPayload) <= len(payload) {
			payload, pflags = altPayload, altFlags
		}
	}

	minT, maxT := int64(0), int64(0)
	if len(bd.ts) > 0 {
		minT = bd.ts[0]
		maxT = bd.ts[len(bd.ts)-1]
	}

	// assemble frame
	frame := binary.LittleEndian.AppendUint64(out, syncMarker)
	frame = binary.LittleEndian.AppendUint32(frame, uint32(blkPreBodySz+len(payload)))
	crcStart := len(frame)
	frame = binary.LittleEndian.AppendUint64(frame, uint64(minT))
	frame = binary.LittleEndian.AppendUint64(frame, uint64(maxT))
	frame = binary.LittleEndian.AppendUint32(frame, uint32(len(bd.ts)))
	frame = append(frame, pflags)
	frame = append(frame, payload...)
	return binary.LittleEndian.AppendUint32(frame, crc32.ChecksumIEEE(frame[crcStart:])), nil
}

func putEncodeScratch(es *encodeScratch) {
	for _, b := range []*[]byte{&es.body, &es.alt, &es.zBody, &es.zAlt} {
		if cap(*b) > maxPooledEncode {
			*b = nil
		}
	}
	if cap(es.fs.w) > maxPooledEncode/8 {
		es.fs = forScratch{}
	}
	encodePool.Put(es)
}

// appendVarint appends v as zig-zag uvarint, like writeVarint.
func appendVarint(dst []byte, v int64) []byte {
	return binary.AppendUvarint(dst, zigzag(v))
}

// appendTimestamps appends the delta-of-delta stream of ts, like encodeTimestamps.
func appendTimestamps(dst []byte, ts []int64) []byte {
	if len(ts) == 0 {
		return dst
	}
	dst = appendVarint(dst, ts[0])
	if len(ts) == 1 {
		return dst
	}
	prevDelta := ts[1] - ts[0]
	dst = appendVarint(dst, prevDelta)
	for i := 2; i < len(ts); i++ {
		delta := ts[i] - ts[i-1]
		dst = appendVarint(dst, delta-prevDelta)
		prevDelta = delta
	}
	return dst
}

// appendValueDeltas appends the delta + zig-zag varint stream of vals, like encodeValues.
func appendValueDeltas(dst []byte, vals []int64) []byte {
	if len(vals) == 0 {
		return dst
	}
	dst = appendVarint(dst, vals[0])
	prev := vals[0]
	for _, v := range vals[1:] {
		dst = appendVarint(dst, v-prev)
		prev = v
	}
	return dst
}

// blockHeader is the cheaply-parsed frame header used to skip or read a block.
type blockHeader struct {
	frameLen  int // total bytes of the frame incl. sync marker and CRC
	minMillis int64
	maxMillis int64
	count     int
	flags     byte
	bodyOff   int // offset within the frame where the compressed/raw body starts
	bodyLen   int // length of the body bytes
}

// parseBlockHeader validates the sync marker and lengths of a frame that starts
// at buf[0]. It does not verify the CRC (done in decodeBlockBody) nor decode the
// body. Returns errCorruptBlock if the frame is malformed or truncated.
func parseBlockHeader(buf []byte) (blockHeader, error) {
	var h blockHeader
	if len(buf) < blkHeaderSz {
		return h, errCorruptBlock
	}
	if binary.LittleEndian.Uint64(buf) != syncMarker {
		return h, errCorruptBlock
	}
	bodyLen := int(binary.LittleEndian.Uint32(buf[blkSync:]))
	if bodyLen < blkPreBodySz {
		return h, errCorruptBlock
	}
	frameLen := blkHeaderSz + bodyLen + blkCRC
	if len(buf) < frameLen {
		return h, errCorruptBlock
	}
	region := buf[blkHeaderSz : blkHeaderSz+bodyLen]
	h.frameLen = frameLen
	h.minMillis = int64(binary.LittleEndian.Uint64(region))
	h.maxMillis = int64(binary.LittleEndian.Uint64(region[8:]))
	h.count = int(binary.LittleEndian.Uint32(region[16:]))
	h.flags = region[20]
	h.bodyOff = blkHeaderSz + blkPreBodySz
	h.bodyLen = bodyLen - blkPreBodySz
	return h, nil
}

// decodeBlockBody verifies the CRC and decodes the body of the frame in buf
// (which must be at least h.frameLen long) into bd. Scratch slices in bd are
// reused (reset to len 0). For enum schemes the ids are decoded; string
// resolution happens in the caller which owns the dictionary.
//
// The hot path is allocation-free: it decodes directly from the (optionally
// decompressed) body byte slice with an integer cursor, using no io.ByteReader
// interface and no per-block temporary. bd.decompBuf is reused for compressed
// blocks so decompression allocates at most once per column read (grow-once).
func decodeBlockBody(buf []byte, h blockHeader, s Scheme, bd *blockData) error {
	if !blockIntact(buf, h) {
		return errCorruptBlock
	}
	return decodeBlockData(buf, h, s, bd)
}

// decodeBlockData is decodeBlockBody without the check of the CRC, for a frame that was checked and unmasked. A flag
// of a later version is errUnknownFormat, a body, which does not decode, errCorruptBlock.
func decodeBlockData(buf []byte, h blockHeader, s Scheme, bd *blockData) error {
	if h.flags&^flagsKnown != 0 {
		return fmt.Errorf("%w: block flags 0x%02x", errUnknownFormat, h.flags)
	}
	body := buf[h.bodyOff : h.bodyOff+h.bodyLen]
	if h.flags&flagCompressed != 0 {
		dec, err := s2.Decode(bd.decompBuf[:cap(bd.decompBuf)], body)
		if err != nil {
			return errCorruptBlock
		}
		bd.decompBuf = dec
		body = dec
	}

	var ok bool
	var off int
	if h.flags&flagEquidistantTS != 0 {
		bd.ts, off, ok = decodeEquidistantTimestampsInto(body, 0, h.count, bd.ts)
	} else {
		bd.ts, off, ok = decodeTimestampsInto(body, 0, h.count, bd.ts)
	}
	if !ok {
		return errCorruptBlock
	}
	constVal := h.flags&flagConstVal != 0
	forVals := h.flags&flagFORVals != 0
	switch s {
	case SchemeDecimal, SchemeEnum:
		switch {
		case constVal && forVals:
			return errCorruptBlock
		case constVal:
			bd.vals, _, ok = decodeConstI64Into(body, off, h.count, bd.vals)
		case forVals:
			bd.vals, _, ok = decodeFORInto(body, off, h.count, bd.vals)
		default:
			bd.vals, _, ok = decodeValuesInto(body, off, h.count, bd.vals)
		}
	case SchemeString:
		if forVals {
			return errCorruptBlock
		}
		bd.strs, _, ok = decodeStringsInto(body, off, h.count, bd.strs)
	default:
		return errInvalidScheme
	}
	if !ok {
		return errCorruptBlock
	}
	return nil
}

// blockIntact reports whether the CRC of the frame in buf (at least h.frameLen
// long) matches, i.e. whether the block was written completely and has not
// changed since. A block torn by a crash fails here even if its length fits.
func blockIntact(buf []byte, h blockHeader) bool {
	crcRegion := buf[blkHeaderSz : blkHeaderSz+blkPreBodySz+h.bodyLen]
	want := binary.LittleEndian.Uint32(buf[h.frameLen-blkCRC : h.frameLen])
	return crc32.ChecksumIEEE(crcRegion) == want
}

func decodeStrings(b *xbytes.Buffer, count int, dst []string) ([]string, error) {
	dst = dst[:0]
	for i := 0; i < count; i++ {
		s, err := b.ReadString()
		if err != nil {
			return dst, err
		}
		dst = append(dst, s)
	}
	return dst, nil
}
