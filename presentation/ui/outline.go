// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"go.wdy.de/nago/presentation/proto"
)

// Outline is a utility component.
// It controls the width, offset and color of an outline.
// Contrary to a border, an outline is not taking any space
// in the page's layout.
type Outline struct {
	Width  int
	Offset int
	Color  Color
	inside bool
}

// ora converts an Outline into its protocol representation for serialization.
func (o Outline) ora() proto.Outline {
	l := proto.Outline{
		Width:  proto.Int(o.Width),
		Offset: proto.Int(o.Offset),
		Color:  proto.Color(o.Color),
	}

	if o.inside {
		l.Offset = proto.Int(-o.Width + o.Offset)
	}

	return l
}

// Inside sets the outline to be inside the element.
func (o Outline) Inside() Outline {
	o.inside = true
	return o
}
