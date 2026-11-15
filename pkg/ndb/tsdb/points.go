// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tsdb

import (
	"encoding/binary"
	"slices"
)

// pointEnc holds points in ascending time order encoded as bytes: the points of the open block, which only the memory
// holds, and the body of a continuation record. Per point:
//
//	| Time   varint | zig-zag: the time of the first point, then the change of the step to the point before
//	| Value  varint | zig-zag: the change to the value before (decimal value or enum id), or
//	| Text          | uvarint length and the bytes (string scheme)
//
// A run of equidistant times takes one byte per point, a slowly changing value one or two; no point takes more than
// 20 bytes besides its text. The buffer grows by an eighth at a time, at most 16 bytes per point, so that the memory
// stays close to the bytes of the points.
type pointEnc struct {
	buf                     []byte
	n                       int
	prevTs, prevStep, prevV int64
	minTs, maxTs            int64
}

func (e *pointEnc) add(scheme Scheme, ts, val int64, str string) {
	var dt uint64
	if e.n == 0 {
		dt = zigzag(ts)
		e.minTs = ts
	} else {
		step := ts - e.prevTs
		dt = zigzag(step - e.prevStep)
		e.prevStep = step
	}
	e.prevTs, e.maxTs = ts, ts
	e.n++
	if scheme == SchemeString {
		var tmp [2 * binary.MaxVarintLen64]byte
		k := binary.PutUvarint(tmp[:], dt)
		k += binary.PutUvarint(tmp[k:], uint64(len(str)))
		e.grow(k + len(str))
		e.buf = append(e.buf, tmp[:k]...)
		e.buf = append(e.buf, str...)
		return
	}
	dv := zigzag(val - e.prevV)
	e.prevV = val
	if cap(e.buf)-len(e.buf) < 2*binary.MaxVarintLen64 {
		e.grow(uvarintLen(dt) + uvarintLen(dv))
	}
	e.buf = binary.AppendUvarint(binary.AppendUvarint(e.buf, dt), dv)
}

func uvarintLen(u uint64) int {
	n := 1
	for u >= 0x80 {
		u >>= 7
		n++
	}
	return n
}

func (e *pointEnc) grow(need int) {
	if cap(e.buf)-len(e.buf) >= need {
		return
	}
	n := len(e.buf) + need
	extra := min(n/8, 16*e.n)
	nb := slices.Grow([]byte(nil), n+extra) // the capacity is all the allocator gives
	e.buf = append(nb, e.buf...)
}

// release forgets the points and gives the memory back.
func (e *pointEnc) release() { *e = pointEnc{} }

// decodePoints appends the count points encoded in body to bd. ok is false if body does not hold exactly them.
func decodePoints(body []byte, count int, scheme Scheme, bd *blockData) bool {
	var prevTs, prevStep, prevV int64
	off := 0
	for i := 0; i < count; i++ {
		u, k := binary.Uvarint(body[off:])
		if k <= 0 {
			return false
		}
		off += k
		var ts int64
		if i == 0 {
			ts = unzigzag(u)
		} else {
			prevStep += unzigzag(u)
			ts = prevTs + prevStep
		}
		prevTs = ts
		bd.ts = append(bd.ts, ts)
		if scheme == SchemeString {
			l, k := binary.Uvarint(body[off:])
			if k <= 0 || uint64(len(body)-off-k) < l {
				return false
			}
			off += k
			bd.strs = append(bd.strs, string(body[off:off+int(l)]))
			off += int(l)
			continue
		}
		u, k = binary.Uvarint(body[off:])
		if k <= 0 {
			return false
		}
		off += k
		prevV += unzigzag(u)
		bd.vals = append(bd.vals, prevV)
	}
	return off == len(body)
}

// encodePoints returns the encoded body of the points of bd from index from on.
func encodePoints(bd *blockData, from int, scheme Scheme) []byte {
	var e pointEnc
	for i := from; i < len(bd.ts); i++ {
		if scheme == SchemeString {
			e.add(scheme, bd.ts[i], 0, bd.strs[i])
		} else {
			e.add(scheme, bd.ts[i], bd.vals[i], "")
		}
	}
	return e.buf
}
