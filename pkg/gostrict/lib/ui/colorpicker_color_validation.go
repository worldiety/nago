// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

import "go.wdy.de/nago/presentation/ui/colorpicker"

// ColorpickerIsValidHexColor reports, whether hex is a color like #RRGGBB or #RRGGBBAA, like
// colorpicker.IsValidHexColor of nago.
//
//gostrict:host
func ColorpickerIsValidHexColor(hex string) bool {
	return colorpicker.IsValidHexColor(hex)
}
