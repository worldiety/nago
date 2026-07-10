// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"math"

	"go.wdy.de/nago/pkg/xbytes"
)

// NA is the not-available sentinel for scaled int64 values, matching the
// convention of pkg/timeseries (math.MinInt64). A hole in time is an absent
// timestamp; an explicitly stored NA is a present timestamp whose value is
// this sentinel.
const NA int64 = math.MinInt64

// scale converts a float64 to a scaled int64 with the given number of decimal
// places, killing float noise per the engine's design. round-half-away-from-zero.
// NaN/Inf map to NA.
func scale(v float64, decimals uint8) int64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return NA
	}
	f := v * pow10(decimals)
	if f >= 0 {
		return int64(f + 0.5)
	}
	return int64(f - 0.5)
}

// unscale converts a scaled int64 back to float64. NA maps to NaN.
func unscale(v int64, decimals uint8) float64 {
	if v == NA {
		return math.NaN()
	}
	return float64(v) / pow10(decimals)
}

var pow10table = [...]float64{
	1e0, 1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9,
	1e10, 1e11, 1e12, 1e13, 1e14, 1e15, 1e16, 1e17, 1e18,
}

func pow10(decimals uint8) float64 {
	if int(decimals) < len(pow10table) {
		return pow10table[decimals]
	}
	return math.Pow10(int(decimals))
}

// zigzag encodes a signed int64 into an unsigned int64 so that small-magnitude
// values (positive or negative) map to small unsigned values, which the
// subsequent uvarint encoding stores compactly.
func zigzag(v int64) uint64 {
	return uint64((v << 1) ^ (v >> 63))
}

// unzigzag reverses zigzag.
func unzigzag(u uint64) int64 {
	return int64(u>>1) ^ -int64(u&1)
}

// writeVarint writes a signed int64 as zig-zag uvarint.
func writeVarint(b *xbytes.Buffer, v int64) {
	_, _ = b.WriteUvarint(zigzag(v))
}

// readVarint reads a zig-zag uvarint signed int64.
func readVarint(b *xbytes.Buffer) (int64, error) {
	u, err := b.ReadUvarint()
	if err != nil {
		return 0, err
	}
	return unzigzag(u), nil
}

// isEquidistant reports whether ts is a run with a single constant step, i.e.
// ts[i]-ts[i-1] is the same for all i. Requires at least two points. When true,
// the block can be stored as (start, step) and reconstructed by formula. A hole
// (any differing delta) makes it false, and the caller falls back to
// delta-of-delta. The step may be any non-negative value (0 only if a block
// somehow held duplicate timestamps, which Fuse prevents; treated as non-equi).
func isEquidistant(ts []int64) (equi bool, step int64) {
	if len(ts) < 2 {
		return false, 0
	}
	step = ts[1] - ts[0]
	if step <= 0 {
		return false, 0
	}
	for i := 2; i < len(ts); i++ {
		if ts[i]-ts[i-1] != step {
			return false, 0
		}
	}
	return true, step
}

// encodeEquidistantTimestamps writes an equidistant timestamp run as two zig-zag
// varints: the start timestamp and the constant step. The reader reconstructs
// every timestamp as start + i*step. This replaces the whole per-point stream
// (~1 byte/point) with a fixed ~10-16 bytes per block.
func encodeEquidistantTimestamps(b *xbytes.Buffer, ts []int64, step int64) {
	writeVarint(b, ts[0])
	writeVarint(b, step)
}

// decodeEquidistantTimestampsInto reconstructs count equidistant timestamps from
// the (start, step) pair at buf[off:] into dst, returning dst, the new offset,
// and ok. Pure arithmetic, no per-point varint.
func decodeEquidistantTimestampsInto(buf []byte, off, count int, dst []int64) (out []int64, next int, ok bool) {
	dst = dst[:0]
	if count == 0 {
		return dst, off, true
	}
	u, off, ok := readUvarintAt(buf, off)
	if !ok {
		return dst, off, false
	}
	start := unzigzag(u)
	if count == 1 {
		return append(dst, start), off, true
	}
	u, off, ok = readUvarintAt(buf, off)
	if !ok {
		return dst, off, false
	}
	step := unzigzag(u)
	t := start
	for i := 0; i < count; i++ {
		dst = append(dst, t)
		t += step
	}
	return dst, off, true
}

// encodeTimestamps writes a delta-of-delta encoded timestamp stream into b.
// For a truly equidistant series every second-order delta is zero, which
// uvarint stores in a single byte (0). ts is assumed strictly ascending.
//
// Layout: count is written by the caller; here we write firstTs (zig-zag
// varint), then for len>=2 the first delta (zig-zag varint), then for the rest
// the delta-of-delta (zig-zag varint).
func encodeTimestamps(b *xbytes.Buffer, ts []int64) {
	if len(ts) == 0 {
		return
	}
	writeVarint(b, ts[0])
	if len(ts) == 1 {
		return
	}
	prevDelta := ts[1] - ts[0]
	writeVarint(b, prevDelta)
	for i := 2; i < len(ts); i++ {
		delta := ts[i] - ts[i-1]
		writeVarint(b, delta-prevDelta)
		prevDelta = delta
	}
}

// decodeTimestamps reads count delta-of-delta encoded timestamps into dst
// (which is reset to length 0 and grown as needed) and returns it.
func decodeTimestamps(b *xbytes.Buffer, count int, dst []int64) ([]int64, error) {
	dst = dst[:0]
	if count == 0 {
		return dst, nil
	}
	first, err := readVarint(b)
	if err != nil {
		return dst, err
	}
	dst = append(dst, first)
	if count == 1 {
		return dst, nil
	}
	prevDelta, err := readVarint(b)
	if err != nil {
		return dst, err
	}
	prev := first + prevDelta
	dst = append(dst, prev)
	for i := 2; i < count; i++ {
		dod, err := readVarint(b)
		if err != nil {
			return dst, err
		}
		prevDelta += dod
		prev += prevDelta
		dst = append(dst, prev)
	}
	return dst, nil
}

// encodeValues writes a plain delta + zig-zag varint encoded int64 value
// stream. Because measured values oscillate with small relative change, the
// deltas are tiny and store in 1-2 bytes. NA values are preserved bit-exact
// (they participate in the delta like any other int64).
func encodeValues(b *xbytes.Buffer, vals []int64) {
	if len(vals) == 0 {
		return
	}
	writeVarint(b, vals[0])
	prev := vals[0]
	for i := 1; i < len(vals); i++ {
		writeVarint(b, vals[i]-prev)
		prev = vals[i]
	}
}

// decodeValues reads count delta+zig-zag encoded int64 values into dst.
func decodeValues(b *xbytes.Buffer, count int, dst []int64) ([]int64, error) {
	dst = dst[:0]
	if count == 0 {
		return dst, nil
	}
	first, err := readVarint(b)
	if err != nil {
		return dst, err
	}
	dst = append(dst, first)
	prev := first
	for i := 1; i < count; i++ {
		d, err := readVarint(b)
		if err != nil {
			return dst, err
		}
		prev += d
		dst = append(dst, prev)
	}
	return dst, nil
}

// ---- allocation-free / interface-free decoders for the read hot path ----
//
// These operate directly on a byte slice with an integer cursor. They avoid the
// xbytes.Buffer value (which escapes to the heap once per block because it is
// passed by pointer into binary.ReadUvarint's io.ByteReader interface) and the
// associated virtual ReadByte call. Everything here is concrete and inlinable.

// readUvarintAt reads a uvarint at buf[off:], returning the value and the new
// offset. ok is false on truncation or overflow.
func readUvarintAt(buf []byte, off int) (v uint64, next int, ok bool) {
	var s uint
	for i := off; i < len(buf); i++ {
		b := buf[i]
		if b < 0x80 {
			if i-off > 9 || (i-off == 9 && b > 1) {
				return 0, off, false // overflow
			}
			return v | uint64(b)<<s, i + 1, true
		}
		v |= uint64(b&0x7f) << s
		s += 7
	}
	return 0, off, false // truncated
}

// decodeTimestampsInto decodes count delta-of-delta timestamps from buf starting
// at off into dst (reset to len 0, grown as needed).
func decodeTimestampsInto(buf []byte, off, count int, dst []int64) (out []int64, next int, ok bool) {
	dst = dst[:0]
	if count == 0 {
		return dst, off, true
	}
	u, off, ok := readUvarintAt(buf, off)
	if !ok {
		return dst, off, false
	}
	first := unzigzag(u)
	dst = append(dst, first)
	if count == 1 {
		return dst, off, true
	}
	u, off, ok = readUvarintAt(buf, off)
	if !ok {
		return dst, off, false
	}
	prevDelta := unzigzag(u)
	prev := first + prevDelta
	dst = append(dst, prev)
	for i := 2; i < count; i++ {
		u, off, ok = readUvarintAt(buf, off)
		if !ok {
			return dst, off, false
		}
		prevDelta += unzigzag(u)
		prev += prevDelta
		dst = append(dst, prev)
	}
	return dst, off, true
}

// decodeValuesInto decodes count delta+zig-zag int64 values from buf at off.
func decodeValuesInto(buf []byte, off, count int, dst []int64) (out []int64, next int, ok bool) {
	dst = dst[:0]
	if count == 0 {
		return dst, off, true
	}
	u, off, ok := readUvarintAt(buf, off)
	if !ok {
		return dst, off, false
	}
	prev := unzigzag(u)
	dst = append(dst, prev)
	for i := 1; i < count; i++ {
		u, off, ok = readUvarintAt(buf, off)
		if !ok {
			return dst, off, false
		}
		prev += unzigzag(u)
		dst = append(dst, prev)
	}
	return dst, off, true
}

// decodeStringsInto decodes count length-prefixed strings from buf at off. The
// strings alias buf (zero-copy) and are only valid while buf is valid.
func decodeStringsInto(buf []byte, off, count int, dst []string) (out []string, next int, ok bool) {
	dst = dst[:0]
	for i := 0; i < count; i++ {
		l, o, uok := readUvarintAt(buf, off)
		if !uok {
			return dst, off, false
		}
		off = o
		if int(l) > len(buf)-off {
			return dst, off, false
		}
		dst = append(dst, string(buf[off:off+int(l)]))
		off += int(l)
	}
	return dst, off, true
}

// ---- constant value stream ----
//
// When every numeric value in a block is identical (a signal parked at a
// setpoint, an idle status, a long unchanged run), the value stream is stored as
// a single zig-zag varint and reconstructed by filling the whole slice. This
// makes a flat block cost O(1) bytes and decode a fill loop instead of a
// per-point varint decode.

// allEqualI64 reports whether all elements of v are identical. Empty and
// single-element slices are trivially constant.
func allEqualI64(v []int64) bool {
	if len(v) < 2 {
		return true
	}
	first := v[0]
	for i := 1; i < len(v); i++ {
		if v[i] != first {
			return false
		}
	}
	return true
}

func firstI64(v []int64) int64 {
	if len(v) == 0 {
		return 0
	}
	return v[0]
}

// encodeConstValue writes a single zig-zag varint holding the constant value.
func encodeConstValue(b *xbytes.Buffer, v int64) {
	writeVarint(b, v)
}

// decodeConstI64Into reads the single constant int64 at buf[off:] and fills dst
// with count copies of it.
func decodeConstI64Into(buf []byte, off, count int, dst []int64) (out []int64, next int, ok bool) {
	dst = dst[:0]
	if count == 0 {
		return dst, off, true
	}
	u, off, ok := readUvarintAt(buf, off)
	if !ok {
		return dst, off, false
	}
	v := unzigzag(u)
	for i := 0; i < count; i++ {
		dst = append(dst, v)
	}
	return dst, off, true
}
