// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"encoding/binary"
	"hash/crc32"

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
//	| Flags       1 | bit0: body is s2-compressed
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

// encodeBlock serializes bd into a self-describing block frame appended to out
// and returns the extended slice. compress controls s2 compression of the body.
func encodeBlock(out []byte, bd *blockData, s Scheme, compress bool) ([]byte, error) {
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

	// encode body (timestamps + values) into a scratch buffer
	body := xbytes.Buffer{}
	if equi {
		encodeEquidistantTimestamps(&body, bd.ts, step)
	} else {
		encodeTimestamps(&body, bd.ts)
	}
	switch s {
	case SchemeDecimal, SchemeEnum:
		if constVal {
			encodeConstValue(&body, firstI64(bd.vals))
		} else {
			encodeValues(&body, bd.vals)
		}
	case SchemeString:
		encodeStrings(&body, bd.strs)
	default:
		return out, errInvalidScheme
	}
	bodyBytes := body.Buf[:body.Pos]

	flags := byte(0)
	if equi {
		flags |= flagEquidistantTS
	}
	if constVal {
		flags |= flagConstVal
	}
	if compress && len(bodyBytes) > 512 {
		c := s2.Encode(nil, bodyBytes)
		if len(c) < len(bodyBytes) {
			bodyBytes = c
			flags |= flagCompressed
		}
	}

	minT, maxT := int64(0), int64(0)
	if len(bd.ts) > 0 {
		minT = bd.ts[0]
		maxT = bd.ts[len(bd.ts)-1]
	}

	// assemble frame
	frame := xbytes.Buffer{Buf: out, Pos: len(out)}
	writeU64(&frame, syncMarker)
	bodyLen := uint32(blkPreBodySz + len(bodyBytes))
	_ = frame.WriteUint32(bodyLen)

	crcStart := frame.Pos
	writeU64(&frame, uint64(minT))
	writeU64(&frame, uint64(maxT))
	_ = frame.WriteUint32(uint32(len(bd.ts)))
	_ = frame.WriteByte(flags)
	_, _ = frame.Write(bodyBytes)
	crc := crc32.ChecksumIEEE(frame.Buf[crcStart:frame.Pos])
	_ = frame.WriteUint32(crc)

	return frame.Buf[:frame.Pos], nil
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
	crcRegion := buf[blkHeaderSz : blkHeaderSz+blkPreBodySz+h.bodyLen]
	want := binary.LittleEndian.Uint32(buf[h.frameLen-blkCRC : h.frameLen])
	if crc32.ChecksumIEEE(crcRegion) != want {
		return errCorruptBlock
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
	switch s {
	case SchemeDecimal, SchemeEnum:
		if constVal {
			bd.vals, _, ok = decodeConstI64Into(body, off, h.count, bd.vals)
		} else {
			bd.vals, _, ok = decodeValuesInto(body, off, h.count, bd.vals)
		}
	case SchemeString:
		bd.strs, _, ok = decodeStringsInto(body, off, h.count, bd.strs)
	default:
		return errInvalidScheme
	}
	if !ok {
		return errCorruptBlock
	}
	return nil
}

func writeU64(b *xbytes.Buffer, v uint64) {
	var tmp [8]byte
	binary.LittleEndian.PutUint64(tmp[:], v)
	_, _ = b.Write(tmp[:])
}

// encodeStrings stores length-prefixed raw strings. The whole block body is
// typically s2-compressed by encodeBlock, so no per-string compression here.
func encodeStrings(b *xbytes.Buffer, strs []string) {
	for _, s := range strs {
		_, _ = b.WriteString(s)
	}
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
