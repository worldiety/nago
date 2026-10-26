// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package rebac

import (
	"encoding/binary"
	"math"
)

type bNamespace uint32
type bInstance uint32
type bRelation uint32

// fixedBinaryTriple is the index representation of a triple: five interned ids, each encoded as a little endian
// word. A zero word denotes a hole, see [DB.scan].
type fixedBinaryTriple [8 + 4 + 8]byte

// lessFixedBinaryTriple orders like bytes.Compare, but compares big endian words instead of single bytes, which
// keeps the btree lookups of the permission checks cheap.
func lessFixedBinaryTriple(a, b fixedBinaryTriple) bool {
	if x, y := binary.BigEndian.Uint64(a[0:8]), binary.BigEndian.Uint64(b[0:8]); x != y {
		return x < y
	}

	if x, y := binary.BigEndian.Uint64(a[8:16]), binary.BigEndian.Uint64(b[8:16]); x != y {
		return x < y
	}

	return binary.BigEndian.Uint32(a[16:20]) < binary.BigEndian.Uint32(b[16:20])
}

func (t fixedBinaryTriple) Unwrap() binaryTriple {
	return binaryTriple{
		A: bEntity{
			Namespace: bNamespace(binary.LittleEndian.Uint32(t[0:4])),
			Instance:  bInstance(binary.LittleEndian.Uint32(t[4:8])),
		},
		Relation: bRelation(binary.LittleEndian.Uint32(t[8:12])),
		B: bEntity{
			Namespace: bNamespace(binary.LittleEndian.Uint32(t[12:16])),
			Instance:  bInstance(binary.LittleEndian.Uint32(t[16:20])),
		},
	}
}

// boundPrefix returns the length in bytes of the leading words, which are not holes. All triples sharing these
// bytes are a single contiguous range in the btree.
func (t fixedBinaryTriple) boundPrefix() int {
	for i := 0; i < len(t); i += 4 {
		if binary.LittleEndian.Uint32(t[i:i+4]) == 0 {
			return i
		}
	}

	return len(t)
}

// afterRelation returns the smallest key, which is larger than every triple sharing the source and the relation
// of t, and false if there is none. It lets a scan skip the remaining block of a relation, e.g. the thousands of
// members of a role. The btree is ordered by the big endian interpretation of each word, thus the next block
// starts at the successor of the relation word.
func (t fixedBinaryTriple) afterRelation() (fixedBinaryTriple, bool) {
	word := binary.BigEndian.Uint32(t[8:12])
	if word == math.MaxUint32 {
		return fixedBinaryTriple{}, false
	}

	var next fixedBinaryTriple
	copy(next[:8], t[:8])
	binary.BigEndian.PutUint32(next[8:12], word+1)
	return next, true
}

type bEntity struct {
	Namespace bNamespace
	Instance  bInstance
}

type binaryTriple struct {
	A        bEntity
	Relation bRelation
	B        bEntity
}

func (t binaryTriple) Binary() fixedBinaryTriple {
	var b fixedBinaryTriple
	binary.LittleEndian.PutUint32(b[0:4], uint32(t.A.Namespace))
	binary.LittleEndian.PutUint32(b[4:8], uint32(t.A.Instance))
	binary.LittleEndian.PutUint32(b[8:12], uint32(t.Relation))
	binary.LittleEndian.PutUint32(b[12:16], uint32(t.B.Namespace))
	binary.LittleEndian.PutUint32(b[16:20], uint32(t.B.Instance))
	return b
}

// reverse swaps A and B.
func (t binaryTriple) reverse() binaryTriple {
	t.A, t.B = t.B, t.A
	return t
}

// matches returns true, if t equals the pattern in every component, which is not a hole in the pattern.
func (t binaryTriple) matches(pattern binaryTriple) bool {
	return (pattern.A.Namespace == 0 || pattern.A.Namespace == t.A.Namespace) &&
		(pattern.A.Instance == 0 || pattern.A.Instance == t.A.Instance) &&
		(pattern.Relation == 0 || pattern.Relation == t.Relation) &&
		(pattern.B.Namespace == 0 || pattern.B.Namespace == t.B.Namespace) &&
		(pattern.B.Instance == 0 || pattern.B.Instance == t.B.Instance)
}
