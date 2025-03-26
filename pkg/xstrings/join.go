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

func If[T ~string](b bool, ifTrue, ifFalse T) T {
	if b {
		return ifTrue
	}

	return ifFalse
}
