// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package localization

import (
	"strings"
	"unicode"
)

// NormalizeAndTitle replaces "_" and "." with spaces,
// collapses multiple spaces, and converts each word
// to Title Case (first letter uppercase, the rest lowercase).
func NormalizeAndTitle(s string) string {
	s = strings.NewReplacer("_", " ", ".", " ").Replace(s)

	parts := strings.Fields(s)

	for i, w := range parts {
		parts[i] = capitalizeWord(w)
	}

	return strings.Join(parts, " ")
}

// capitalizeWord lowercases the entire word and
// then makes the first rune uppercase (Unicode-safe).
func capitalizeWord(w string) string {
	r := []rune(strings.ToLower(w))
	if len(r) == 0 {
		return w
	}
	r[0] = unicode.ToTitle(r[0])
	return string(r)
}
