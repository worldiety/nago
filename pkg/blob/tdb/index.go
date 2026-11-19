// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tdb

import (
	"strings"
	"unsafe"

	"github.com/google/btree"
)

// item is the entry of the in-memory index. It takes 32 bytes, its key lives in a [keyArena].
type item struct {
	key string
	off uint64 // offset of the value within its file, the flag inLog tells the file
	len uint32
}

// inLog marks the offset of a value, which lives in the log instead of the compacted file.
const inLog = 1 << 63

// treeDegree is the degree of the bucket trees. A higher degree wastes less memory per entry, a lower degree copies
// less when the first write after a snapshot of the iterators clones a path of the tree.
const treeDegree = 32

func newTree() *btree.BTreeG[item] {
	return btree.NewG[item](treeDegree, func(a, b item) bool {
		return a.key < b.key
	})
}

// buildTree builds the tree of the items, of which a later one replaces an earlier one with the same key.
//
// Inserting sorted items one after another leaves every node half full, because a full node splits in the middle and
// the left half never gets another item. Thus, strictly ascending items are inserted in two passes: every second item
// first, which leaves half full nodes with gaps, and then the others, which fill the gaps. The nodes end up nearly
// full, which halves the memory of the tree.
func buildTree(items []item) *btree.BTreeG[item] {
	t := newTree()
	for i := 1; i < len(items); i++ {
		if items[i-1].key >= items[i].key {
			for _, it := range items {
				t.ReplaceOrInsert(it)
			}

			return t
		}
	}

	for i := 0; i < len(items); i += 2 {
		t.ReplaceOrInsert(items[i])
	}

	for i := 1; i < len(items); i += 2 {
		t.ReplaceOrInsert(items[i])
	}

	return t
}

// arenaChunk is the size of the blocks of a keyArena. A key takes a quarter of its size at most.
const arenaChunk = 64 * 1024

// keyArena copies keys into large blocks. A separate allocation per key costs its size class (48 bytes for a UUID of
// 36 bytes) and an object, which the garbage collector has to mark. A block lives as long as one of its keys, thus the
// keys of removed entries stay until the next compaction builds a new arena.
type keyArena struct {
	chunk []byte
}

// add returns a copy of s, which may refer to a temporary buffer.
func (a *keyArena) add(s string) string {
	if len(s) == 0 {
		return ""
	}

	if len(s) > arenaChunk/4 {
		return strings.Clone(s)
	}

	if cap(a.chunk)-len(a.chunk) < len(s) {
		a.chunk = make([]byte, 0, arenaChunk)
	}

	start := len(a.chunk)
	a.chunk = append(a.chunk, s...)
	// the bytes up to len(a.chunk) are never written again, thus the string stays immutable
	return unsafe.String(&a.chunk[start], len(s))
}

// unsafeString returns the bytes as string without a copy. It is only valid as long as b is not modified.
func unsafeString(b []byte) string {
	if len(b) == 0 {
		return ""
	}

	return unsafe.String(&b[0], len(b))
}
