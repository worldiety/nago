// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package xslices contains some proposals from https://github.com/golang/go/issues/61899.
// This package will be removed, as soon as these functions become available.
package xslices

import (
	"iter"
	"sort"
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

// PrefixSearch takes a sorted slice of strings and a prefix, and returns the subslice
// containing all strings that start with the given prefix.
//
// The function uses binary search to find the range of matching elements efficiently:
//  1. Find the lowest index where the prefix could appear.
//  2. Find the upper bound (the point where strings no longer start with the prefix).
//
// Complexity: O(log n + k), where n is the length of the slice and k is the number of matches.
func PrefixSearch[Slice ~[]Elem, Elem ~string](data Slice, prefix Elem) Slice {
	// Find the lower bound: the first index where prefix could appear
	start := sort.Search(len(data), func(i int) bool {
		return data[i] >= prefix
	})

	// Construct an artificial upper bound (prefix + high Unicode character)
	upperBound := prefix + "\uffff"

	// Find the upper bound: the first index where elements are greater than prefix range
	end := sort.Search(len(data), func(i int) bool {
		return data[i] >= upperBound
	})

	// Return the subslice
	return data[start:end]
}
