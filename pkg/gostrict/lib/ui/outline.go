// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//gostrict:header
package ui

// OutlineStyle is the style of the line of an outline.
type OutlineStyle int32

const (
	OutlineSolid  OutlineStyle = 0
	OutlineDotted OutlineStyle = 1
)

// Outline is a line around a view with its width, offset and color. Contrary to a Border, an outline takes no space
// in the layout. See TStack.Outline.
type Outline struct {
	Style OutlineStyle
	// Width is the thickness of the line.
	Width int32
	// Offset is the distance between the view and the line.
	Offset int32
	Color  Color
}
