// Package slices contains some proposals from https://github.com/golang/go/issues/61899.
// This package will be removed, as soon as these functions become available.
package xslices

import (
	"iter"
)

// Values2 returns an iterator over the values in the slice,
// starting with s[0].
func Values2[Slice ~[]Elem, Elem, T any](s Slice) iter.Seq2[Elem, T] {
	return func(yield func(Elem, T) bool) {
		var zero T
		for _, v := range s {
			if !yield(v, zero) {
				return
			}
		}
	}
}

func Collect2[Slice ~[]Elem, Elem any](s iter.Seq2[Elem, error]) (Slice, error) {
	var res Slice
	for elem, e := range s {
		if e != nil {
			return res, e
		}
		res = append(res, elem)
	}

	return res, nil
}
