// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package xstrings

import "testing"

func TestCompareFold(t *testing.T) {
	tests := []struct {
		a, b   string
		expect int
	}{
		{"abc", "abc", 0},
		{"abc", "ABC", 0},
		{"AbC", "aBc", 0},
		{"abc", "abcd", -1},
		{"abcd", "abc", 1},
		{"banana", "apple", 1},
		{"apple", "Banana", -1},
		{"Äpfel", "äpfel", 0}, // Unicode
		{"ß", "SS", 1},        // Unicode edge case
		{"hello", "helloo", -1},
		{"helloo", "hello", 1},
		{"", "", 0},
		{"", "nonempty", -1},
		{"nonempty", "", 1},
	}

	for _, tt := range tests {
		got := CompareIgnoreCase(tt.a, tt.b)
		if got != tt.expect {
			t.Errorf("CompareFold(%q, %q) = %d; want %d", tt.a, tt.b, got, tt.expect)
		}
	}
}
