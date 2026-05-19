// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package colorpicker

// IsValidHexColor validates CSS-like hex colors: #RGB, #RGBA, #RRGGBB, #RRGGBBAA.
// Returns true for valid hex color strings, false otherwise.
func IsValidHexColor(s string) bool {
	if len(s) < 4 || s[0] != '#' {
		return false
	}
	n := len(s) - 1
	if n != 3 && n != 4 && n != 6 && n != 8 {
		return false
	}
	for i := 1; i < len(s); i++ {
		if !isHexDigit(s[i]) {
			return false
		}
	}
	return true
}

func isHexDigit(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}
