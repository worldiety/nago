package maps

import (
	"cmp"
	"go.wdy.de/nago/pkg/iter"
	"slices"
)

func Keys[K comparable, V any](m map[K]V) iter.Seq[K] {
	return func(yield func(K) bool) {
		for k := range m {
			if !yield(k) {
				return
			}
		}
	}
}

// SortedKeys returns all keys in its natural sorted order. Probably there will never be something in the std lib.
func SortedKeys[K cmp.Ordered, V any](m map[K]V) []K {
	tmp := make([]K, 0, len(m))
	for k := range m {
		tmp = append(tmp, k)
	}

	slices.Sort(tmp)
	return tmp
}

func Values[K comparable, V any](m map[K]V) iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, v := range m {
			if !yield(v) {
				return
			}
		}
	}
}
