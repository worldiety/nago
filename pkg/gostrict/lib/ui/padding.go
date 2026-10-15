// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//gostrict:header
package ui

import nui "go.wdy.de/nago/presentation/ui"

// Padding is the space inside a view at each side.
type Padding struct {
	Top    Length
	Left   Length
	Right  Length
	Bottom Length
}

// All sets the same padding at each side, e.g. ui.Padding{}.All(ui.L16).
//
//gostrict:host
func (p Padding) All(pad Length) Padding {
	return paddingFrom(p.nago().All(nui.Length(pad)))
}

// Vertical sets the padding at the top and the bottom.
//
//gostrict:host
func (p Padding) Vertical(pad Length) Padding {
	return paddingFrom(p.nago().Vertical(nui.Length(pad)))
}

// Horizontal sets the padding at the left and the right.
//
//gostrict:host
func (p Padding) Horizontal(pad Length) Padding {
	return paddingFrom(p.nago().Horizontal(nui.Length(pad)))
}
