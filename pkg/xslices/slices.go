// Package xslices contains some proposals from https://github.com/golang/go/issues/61899.
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

// Collect2 collects until Seq2 finds an error and returns it and the collected values.
func Collect2[E any](s iter.Seq2[E, error]) ([]E, error) {
	var res []E
	for elem, err := range s {
		if err != nil {
			return res, err
		}

		res = append(res, elem)
	}

	return res, nil
}

// ValuesWithError creates an iter.Seq2 which either yields one err if not nil and otherwise yields all slice elements.
func ValuesWithError[Slice ~[]Elem, Elem any](s Slice, err error) iter.Seq2[Elem, error] {
	return func(yield func(Elem, error) bool) {
		var zero Elem
		if err != nil {
			yield(zero, err)
			return
		}

		for _, v := range s {
			if !yield(v, nil) {
				return
			}
		}
	}
}
