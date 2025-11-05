// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package xstrings

import (
	"unicode"
	"unicode/utf8"
)

func CompareIgnoreCase(a, b string) int {
	for len(a) > 0 && len(b) > 0 {
		ra, sizeA := utf8.DecodeRuneInString(a)
		rb, sizeB := utf8.DecodeRuneInString(b)
		la := unicode.ToLower(ra)
		lb := unicode.ToLower(rb)
		if la != lb {
			if la < lb {
				return -1
			}
			return 1
		}
		a = a[sizeA:]
		b = b[sizeB:]
	}
	switch {
	case len(a) == 0 && len(b) == 0:
		return 0
	case len(a) == 0:
		return -1
	default:
		return 1
	}
}
