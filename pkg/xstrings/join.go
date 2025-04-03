// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package xstrings

import (
	"strings"
)

func Join[T ~string](s []T, sep string) T {
	tmp := make([]string, 0, len(s))
	for i := range s {
		tmp = append(tmp, string(s[i]))
	}

	return T(strings.Join(tmp, sep))
}

func Join2[T ~string](sep, a, b T) T {
	if a == "" {
		return b
	}

	if b == "" {
		return a
	}

	return a + sep + b
}

// Space concats all given values with white space. If any value is empty, it is discarded and no whitespace is
// added.
func Space[T ~string](values ...T) T {
	if len(values) == 0 {
		return ""
	}

	if len(values) == 1 {
		return values[0]
	}

	if len(values) == 2 {
		return Join2(" ", values[0], values[1])
	}

	tmp := make([]string, 0, len(values))
	for _, v := range values {
		if v != "" {
			tmp = append(tmp, string(v))
		}
	}

	return T(strings.Join(tmp, " "))
}

func If[T ~string](b bool, ifTrue, ifFalse T) T {
	if b {
		return ifTrue
	}

	return ifFalse
}
