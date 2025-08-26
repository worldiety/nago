// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"fmt"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

// Border defines the border and shadow styling for a component (Border).
// It controls the radius of each corner, the width and color of each edge,
// and optional shadow settings. A border affects the component's layout
// dimensions because the width is applied outside the content box.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Border struct {
	TopLeftRadius     Length
	TopRightRadius    Length
	BottomLeftRadius  Length
	BottomRightRadius Length

	LeftWidth   Length
	TopWidth    Length
	RightWidth  Length
	BottomWidth Length

	LeftColor   Color
	TopColor    Color
	RightColor  Color
	BottomColor Color

	BoxShadow Shadow `json:"s,omitempty"`
}

// ora converts a Border into its protocol representation for serialization.
func (b Border) ora() proto.Border {
	return proto.Border{
		TopLeftRadius:     proto.Length(b.TopLeftRadius),
		TopRightRadius:    proto.Length(b.TopRightRadius),
		BottomLeftRadius:  proto.Length(b.BottomLeftRadius),
		BottomRightRadius: proto.Length(b.BottomRightRadius),
		LeftWidth:         proto.Length(b.LeftWidth),
		TopWidth:          proto.Length(b.TopWidth),
		RightWidth:        proto.Length(b.RightWidth),
		BottomWidth:       proto.Length(b.BottomWidth),
		LeftColor:         proto.Color(b.LeftColor),
		TopColor:          proto.Color(b.TopColor),
		RightColor:        proto.Color(b.RightColor),
		BottomColor:       proto.Color(b.BottomColor),
		BoxShadow:         b.BoxShadow.ora(),
	}
}

// Radius sets the same corner radius for all four corners.
func (b Border) Radius(radius Length) Border {
	b.TopLeftRadius = radius
	b.TopRightRadius = radius
	b.BottomLeftRadius = radius
	b.BottomRightRadius = radius
	return b
}

// TopRadius sets the same corner radius for both top corners.
func (b Border) TopRadius(radius Length) Border {
	b.TopLeftRadius = radius
	b.TopRightRadius = radius
	return b
}

// BottomRadius sets the same corner radius for both bottom corners.
func (b Border) BottomRadius(radius Length) Border {
	b.BottomLeftRadius = radius
	b.BottomRightRadius = radius
	return b
}

// Circle sets all corner radii to a large value, resulting in a circle/ellipse shape.
func (b Border) Circle() Border {
	return b.Radius("999999dp")
}

// Width sets the same border thickness on all four sides.
func (b Border) Width(width Length) Border {
	b.LeftWidth = width
	b.TopWidth = width
	b.RightWidth = width
	b.BottomWidth = width
	return b
}

// Color sets the same border color on all four sides.
func (b Border) Color(c Color) Border {
	b.LeftColor = c
	b.TopColor = c
	b.BottomColor = c
	b.RightColor = c
	return b
}

// Shadow adds a shadow with the given blur radius and a default semi-transparent black color.
func (b Border) Shadow(radius Length) Border {
	b.BoxShadow.Radius = radius
	b.BoxShadow.Color = "#00000054"
	b.BoxShadow.X = ""
	b.BoxShadow.Y = ""
	return b
}

// Elevate applies a shadow effect based on a given elevation in device-independent pixels (DP).
// The shadow blur radius and offset are scaled accordingly, using rem units.
func (b Border) Elevate(elevation core.DP) Border {
	rem := float64(elevation) / 16
	b.BoxShadow.Radius = Length(fmt.Sprintf("%.2frem", rem*3))
	b.BoxShadow.Color = "#00000030"
	b.BoxShadow.X = ""
	b.BoxShadow.Y = Length(fmt.Sprintf("%.2frem", rem))
	return b
}
