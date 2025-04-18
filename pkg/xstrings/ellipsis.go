// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package xstrings

import (
	"strings"
	"unicode/utf8"
)

// EllipsisEnd returns either s or a truncated version ending with ...
// N is the max rune count, which the returned text contains.
func EllipsisEnd[Str ~string](s Str, n int) Str {
	if utf8.RuneCountInString(string(s)) <= n {
		return s
	}

	n -= 3

	var sb strings.Builder
	for idx, r := range s {
		if idx < n {
			sb.WriteRune(r)
		} else {
			break
		}
	}

	sb.WriteString("...")
	return Str(sb.String())
}
