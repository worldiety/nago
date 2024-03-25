// Package slices contains some proposals from https://github.com/golang/go/issues/61899
package slices

import "go.wdy.de/nago/pkg/iter"

// Values returns an iterator over the values in the slice,
// starting with s[0].
func Values[Slice ~[]Elem, Elem any](s Slice) iter.Seq[Elem] {
	return func(yield func(Elem) bool) {
		for _, v := range s {
			if !yield(v) {
				return
			}
		}
	}
}

// Append appends the values from seq to the slice and returns the extended slice.
func Append[Slice ~[]Elem, Elem any](x Slice, seq iter.Seq[Elem]) Slice {
	seq(func(elem Elem) bool {
		x = append(x, elem)
		return true
	})

	return x
}

// Collect collects values from seq into a new slice and returns it.
func Collect[Elem any](seq iter.Seq[Elem]) []Elem {
	return Append([]Elem(nil), seq)
}
