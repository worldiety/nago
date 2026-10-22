// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package rebac

import "encoding/binary"

type bNamespace uint32
type bInstance uint32
type bRelation uint32

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
