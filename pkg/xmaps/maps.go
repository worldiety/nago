package xmaps

import (
	"cmp"
	"maps"
	"slices"
)

// SortedKeys returns all keys in its natural sorted order. Probably there will never be something in the std lib.
func SortedKeys[K cmp.Ordered, V any](m map[K]V) []K {
	tmp := slices.Collect(maps.Keys(m))
	slices.Sort(tmp)
	return tmp
}
