// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package colorpicker

import "go.wdy.de/nago/pkg/xcolor"

// IsValidHexColor validates CSS-like hex colors using xcolor.ParseHex.
// Accepts #RRGGBB and #RRGGBBAA; returns true for valid hex color strings.
func IsValidHexColor(hex string) bool {
	_, err := xcolor.ParseHex(hex)
	return err == nil
}
