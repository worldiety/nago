// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//gostrict:header
package ui

import (
	"go.wdy.de/nago/presentation/core"
	nui "go.wdy.de/nago/presentation/ui"
)

// BorderStyle is the style of the lines of a border.
type BorderStyle int32

const (
	BorderStyleSolid  BorderStyle = 0
	BorderStyleDotted BorderStyle = 1
	BorderStyleDashed BorderStyle = 2
)

// Border is the border of a view with its radius, width and color at each side and its shadow.
type Border struct {
	TopLeftRadius     Length
	TopRightRadius    Length
	BottomLeftRadius  Length
	BottomRightRadius Length
	LeftWidth         Length
	TopWidth          Length
	RightWidth        Length
	BottomWidth       Length
	LeftColor         Color
	TopColor          Color
	RightColor        Color
	BottomColor       Color
	BoxShadow         Shadow
	BorderStyle       BorderStyle
}

// Shadow is the shadow of a view.
type Shadow struct {
	Color  Color
	Radius Length
	X      Length
	Y      Length
}

// Radius sets the radius of each corner.
//
//gostrict:host
func (b Border) Radius(radius Length) Border {
	return borderFrom(b.nago().Radius(nui.Length(radius)))
}

// Width sets the width of each side.
//
//gostrict:host
func (b Border) Width(width Length) Border {
	return borderFrom(b.nago().Width(nui.Length(width)))
}

// Color sets the color of each side.
//
//gostrict:host
func (b Border) Color(c Color) Border {
	return borderFrom(b.nago().Color(nui.Color(c)))
}

// Elevate sets a shadow, which lifts the view by the elevation in dp.
//
//gostrict:host
func (b Border) Elevate(elevation float64) Border {
	return borderFrom(b.nago().Elevate(core.DP(elevation)))
}

// TopRadius sets the same corner radius for both top corners.
//
//gostrict:host
func (b Border) TopRadius(radius Length) Border {
	return borderFrom(b.nago().TopRadius(nui.Length(radius)))
}

// BottomRadius sets the same corner radius for both bottom corners.
//
//gostrict:host
func (b Border) BottomRadius(radius Length) Border {
	return borderFrom(b.nago().BottomRadius(nui.Length(radius)))
}

// Circle sets all corner radii to a large value, which makes the view a circle or an ellipse.
//
//gostrict:host
func (b Border) Circle() Border {
	return borderFrom(b.nago().Circle())
}

// Style sets the style of the lines, e.g. BorderStyleDashed.
//
//gostrict:host
func (b Border) Style(style BorderStyle) Border {
	return borderFrom(b.nago().Style(nui.BorderStyle(style)))
}

// Shadow adds a shadow with the blur radius and a semi-transparent black color. An empty radius removes the shadow.
//
//gostrict:host
func (b Border) Shadow(radius Length) Border {
	return borderFrom(b.nago().Shadow(nui.Length(radius)))
}
