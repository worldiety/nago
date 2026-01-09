// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package xmaps

import (
	"cmp"
	"iter"
	"maps"
	"slices"
)

// SortedKeys returns all keys in its natural sorted order. Probably there will never be something in the std lib.
func SortedKeys[K cmp.Ordered, V any](m map[K]V) []K {
	tmp := slices.Collect(maps.Keys(m))
	slices.Sort(tmp)
	return tmp
}

// All loops in a deterministic way over the given map.
func All[K cmp.Ordered, V any](m map[K]V) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for _, k := range SortedKeys(m) {
			if !yield(k, m[k]) {
				return
			}
		}
	}
}
