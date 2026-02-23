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

	"go.wdy.de/nago/pkg/cloner"
)

type MutMap[K comparable, V any] interface {
	Map[K, V]
	Put(K, V)
}

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

func Clone[K comparable, V cloner.Cloner[V]](m map[K]V) map[K]V {
	var clone = make(map[K]V, len(m))
	for k, v := range m {
		clone[k] = v.Clone()
	}

	return clone
}
