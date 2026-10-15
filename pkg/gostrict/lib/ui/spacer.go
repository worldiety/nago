// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

import (
	"go.wdy.de/nago/pkg/gostrict/std/js"
	nui "go.wdy.de/nago/presentation/ui"
)

// TSpace is a fixed-size spacer, see Space. Its width and height are both the size.
type TSpace struct{ ref js.Ref }

func (TSpace) isView() {}

// Space returns a fixed-size spacer with the given length, which adds consistent spacing between views.
//
//gostrict:host
func Space(size Length) TSpace {
	return TSpace{ref: wrap(nui.Space(nui.Length(size)))}
}

// Size sets the size of the spacer, which is applied to both width and height.
//
//gostrict:host
func (t TSpace) Size(size Length) TSpace {
	return TSpace{ref: wrap(get[nui.TSpace](t.ref).Size(nui.Length(size)))}
}

// TSpacer is a dynamic spacer, see Spacer. It grows and shrinks to fill the available space inside a VStack or an
// HStack.
type TSpacer struct{ ref js.Ref }

func (TSpacer) isView() {}

// Spacer returns a dynamic spacer, which expands to fill the available space.
//
//gostrict:host
func Spacer() TSpacer {
	return TSpacer{ref: wrap(nui.Spacer())}
}

// BackgroundColor sets the background color of the spacer.
//
//gostrict:host
func (c TSpacer) BackgroundColor(backgroundColor Color) TSpacer {
	return TSpacer{ref: wrap(get[nui.TSpacer](c.ref).BackgroundColor(nui.Color(backgroundColor)))}
}

// Frame sets the layout frame of the spacer, which controls its layout constraints.
//
//gostrict:host
func (c TSpacer) Frame(frame Frame) TSpacer {
	return TSpacer{ref: wrap(get[nui.TSpacer](c.ref).Frame(frame.nago()))}
}
