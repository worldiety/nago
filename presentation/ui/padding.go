// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import "go.wdy.de/nago/presentation/proto"

// Padding is a util component (Padding).
// It defines the spacing inside a component on each side (top, left, right, bottom).
type Padding struct {
	Top    Length // padding at the top
	Left   Length // padding at the left
	Right  Length // padding at the right
	Bottom Length // padding at the bottom
}

// ora converts the padding to its protocol representation.
func (p Padding) ora() proto.Padding {
	return proto.Padding{
		Top:    proto.Length(p.Top),
		Left:   proto.Length(p.Left),
		Right:  proto.Length(p.Right),
		Bottom: proto.Length(p.Bottom),
	}
}

// All applies the same padding value to all four sides.
func (p Padding) All(pad Length) Padding {
	p.Left = pad
	p.Right = pad
	p.Bottom = pad
	p.Top = pad
	return p
}

// Vertical sets the padding for the vertical axis (top and bottom).
func (p Padding) Vertical(pad Length) Padding {
	p.Bottom = pad
	p.Top = pad
	return p
}

// Horizontal sets the padding for the horizontal axis (left and right).
func (p Padding) Horizontal(pad Length) Padding {
	p.Left = pad
	p.Right = pad
	return p
}
