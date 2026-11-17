// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"encoding/binary"
	"math/bits"
	"slices"
)

//go:generate go run ./internal/unpackgen unpack_gen.go

// The frame-of-reference value stream of a block with flagFORVals: the deltas of the values, packed with a fixed bit
// width per sub-block of forSub values, and the positions of the NA values in a list in front of them, so that a
// rare NA does not widen a sub-block to 64 bits. A block holds it instead of the varint stream, if it is not larger
// (see encodeBlock); it decodes a value with a shift, a mask and two additions instead of a varint:
//
//	| NACount  uvarint | number of NA values
//	| NAGap    uvarint | per NA value: its position minus the one of the NA value before minus 1 (the first: its position)
//	| Start    varint  | zig-zag: the value before the first delta
//	per sub-block of forSub values (the last one may be shorter):
//	| Base     varint  | zig-zag: the smallest delta of the sub-block
//	| Width    1       | bits per delta, 0..64
//	| Words    8*Width per group of 64 deltas, little endian; delta j of a group minus Base lies in the bits
//	                   [j*Width, (j+1)*Width) of its words; a short last group is padded with zeros
//
// In the delta stream an NA value repeats the value before it (or the first value, which is not NA), so that the
// deltas around it stay small; the reader puts NA back at its positions. The first delta equals the second one, so
// that the first value does not widen its sub-block. Deltas wrap around like int64 additions, so that every int64
// value, also math.MinInt64 and math.MaxInt64 next to each other, round-trips.
const (
	forSub   = 128 // values per sub-block of one bit width
	forGroup = 64  // values per unpacker call
)

// forScratch holds the buffers of the encoder.
type forScratch struct {
	w []int64  // the values with NA replaced
	d []int64  // the deltas
	u []uint64 // the deltas of a sub-block minus its base
}

// appendFORValues appends the frame-of-reference stream of vals to dst.
func appendFORValues(dst []byte, vals []int64, sc *forScratch) []byte {
	n := len(vals)
	nna := 0
	for _, v := range vals {
		if v == NA {
			nna++
		}
	}
	dst = binary.AppendUvarint(dst, uint64(nna))
	w := vals
	if nna > 0 {
		sc.w = slices.Grow(sc.w[:0], n)[:n]
		w = sc.w
		fill := int64(0)
		for _, v := range vals {
			if v != NA {
				fill = v
				break
			}
		}
		last := -1
		for i, v := range vals {
			if v == NA {
				dst = binary.AppendUvarint(dst, uint64(i-last-1))
				last = i
				w[i] = fill
				continue
			}
			w[i] = v
			fill = v
		}
	}
	if n == 0 {
		return binary.AppendUvarint(dst, 0)
	}

	sc.d = slices.Grow(sc.d[:0], n)[:n]
	d := sc.d
	for i := 1; i < n; i++ {
		d[i] = w[i] - w[i-1]
	}
	if n > 1 {
		d[0] = d[1]
	} else {
		d[0] = 0
	}
	dst = binary.AppendUvarint(dst, zigzag(w[0]-d[0]))

	sc.u = slices.Grow(sc.u[:0], forSub)[:forSub]
	for s := 0; s < n; s += forSub {
		e := min(s+forSub, n)
		mn, mx := d[s], d[s]
		for _, x := range d[s+1 : e] {
			mn = min(mn, x)
			mx = max(mx, x)
		}
		width := bits.Len64(uint64(mx) - uint64(mn))
		dst = binary.AppendUvarint(dst, zigzag(mn))
		dst = append(dst, byte(width))
		u := sc.u[:e-s]
		for j := range u {
			u[j] = uint64(d[s+j]) - uint64(mn)
		}
		dst = appendPacked(dst, u, width)
	}
	return dst
}

// appendPacked appends the values of u with width bits each in groups of 64, each group in width words.
func appendPacked(dst []byte, u []uint64, width int) []byte {
	if width == 0 {
		return dst
	}
	groups := (len(u) + forGroup - 1) / forGroup
	at := len(dst)
	dst = slices.Grow(dst, groups*width*8)[:at+groups*width*8]
	var acc uint64
	var used uint // bits of acc in use
	for g := range groups {
		for j := range forGroup {
			var x uint64
			if k := g*forGroup + j; k < len(u) {
				x = u[k]
			}
			acc |= x << used
			if used+uint(width) >= 64 {
				binary.LittleEndian.PutUint64(dst[at:], acc)
				at += 8
				acc = x >> (64 - used) // 0 for used 0: a shift by 64 is 0 in Go
				used = used + uint(width) - 64
				continue
			}
			used += uint(width)
		}
	}
	return dst
}

// forCap is the capacity, which decodeFORInto needs for count values: it unpacks whole groups.
func forCap(count int) int {
	return (count + forGroup - 1) &^ (forGroup - 1)
}

// decodeFORInto decodes the count values of the frame-of-reference stream at buf[off:] into dst, whose capacity it
// grows to forCap(count), because it writes the values of a short last group beyond count. ok is false if buf does
// not hold such a stream.
func decodeFORInto(buf []byte, off, count int, dst []int64) (out []int64, next int, ok bool) {
	if count > forSub/2*(len(buf)-off) {
		// a sub-block takes at least 2 bytes: buf can not hold count values
		return dst[:0], off, false
	}
	if cap(dst) < forCap(count) {
		dst = make([]int64, forCap(count))
	}
	dst = dst[:forCap(count)]
	nna, off, ok := readUvarintAt(buf, off)
	if !ok || nna > uint64(count) {
		return dst[:0], off, false
	}
	naAt := off
	for range nna {
		if _, off, ok = readUvarintAt(buf, off); !ok {
			return dst[:0], off, false
		}
	}
	if count == 0 {
		_, off, ok = readUvarintAt(buf, off)
		return dst[:0], off, ok
	}
	u, off, ok := readUvarintAt(buf, off)
	if !ok {
		return dst[:0], off, false
	}
	prev := unzigzag(u)
	for s := 0; s < count; s += forSub {
		e := min(s+forSub, count)
		var bu uint64
		if bu, off, ok = readUvarintAt(buf, off); !ok || off >= len(buf) {
			return dst[:0], off, false
		}
		base := unzigzag(bu)
		width := int(buf[off])
		off++
		if width > 64 {
			return dst[:0], off, false
		}
		unpack := unpackDelta[width]
		size := 8 * width
		for j := s; j < e; j += forGroup {
			if off+size > len(buf) {
				return dst[:0], off, false
			}
			prev = unpack(buf[off:off+size], base, prev, (*[forGroup]int64)(dst[j:j+forGroup]))
			off += size
		}
	}
	pos := -1
	for range nna {
		var gap uint64
		gap, naAt, _ = readUvarintAt(buf, naAt)
		if gap >= uint64(count) {
			return dst[:0], off, false
		}
		pos += int(gap) + 1
		if pos >= count {
			return dst[:0], off, false
		}
		dst[pos] = NA
	}
	return dst[:count], off, true
}
