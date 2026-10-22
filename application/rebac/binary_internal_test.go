// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package rebac

import (
	"bytes"
	"math/rand/v2"
	"testing"
)

// TestLessFixedBinaryTripleOrdersLikeBytes ensures that the btree order did not change, because range queries
// rely on prefixes being contiguous.
func TestLessFixedBinaryTripleOrdersLikeBytes(t *testing.T) {
	rnd := rand.New(rand.NewPCG(1, 2))

	// mostly small values with many zeros, like interned pointers, so that a lot of pairs share long prefixes
	random := func() fixedBinaryTriple {
		var b fixedBinaryTriple
		for i := range b {
			if rnd.IntN(3) == 0 {
				b[i] = byte(rnd.IntN(3))
			}
		}

		return b
	}

	for range 100_000 {
		a, b := random(), random()
		if rnd.IntN(4) == 0 {
			b = a
		}

		want := bytes.Compare(a[:], b[:]) < 0
		if got := lessFixedBinaryTriple(a, b); got != want {
			t.Fatalf("less(%v, %v): got %v, want %v", a, b, got, want)
		}
	}
}
