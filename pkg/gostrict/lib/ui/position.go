// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//gostrict:header
package ui

// PositionType is the way, how a view is positioned, see Position.
type PositionType int32

const (
	// PositionDefault is the default, any explicit position value has no effect.
	PositionDefault PositionType = 0
	// PositionOffset is like PositionDefault, but moves the view by the position values after its layout.
	PositionOffset PositionType = 1
	// PositionAbsolute removes the view from the layout and places it by the position values within its parent.
	PositionAbsolute PositionType = 2
	// PositionFixed removes the view from the layout and places it at a fixed position of the viewport.
	PositionFixed PositionType = 3
	// PositionSticky is for completion, and it is unclear, which rules to follow on mobile clients.
	PositionSticky PositionType = 4
)

// PositionRelative is the same as PositionOffset.
const PositionRelative PositionType = 1

// Position places a view within its parent layout. See TStack.Position.
type Position struct {
	Type PositionType
	// Left defines the length from within the left border of the parent.
	Left Length
	// Top defines the length from within the top border of the parent.
	Top Length
	// Right defines the length seen from the right border of the parent.
	Right Length
	// Bottom defines the length seen from the bottom border of the parent.
	Bottom Length
	// ZIndex influences the drawing order of positioned views, a higher index is drawn later, so on top of others.
	ZIndex int32
}
