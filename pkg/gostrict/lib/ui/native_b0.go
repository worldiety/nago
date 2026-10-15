// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package ui

import (
	"go.wdy.de/nago/presentation/core"
	nui "go.wdy.de/nago/presentation/ui"
)

// The conversions of batch b0, see native.go.

func (s IntState) nago() *core.State[int64] {
	return get[*core.State[int64]](s.ref)
}

func (s FloatState) nago() *core.State[float64] {
	return get[*core.State[float64]](s.ref)
}

func (p Position) nago() nui.Position {
	return nui.Position{Type: nui.PositionType(p.Type), Left: nui.Length(p.Left), Top: nui.Length(p.Top), Right: nui.Length(p.Right), Bottom: nui.Length(p.Bottom), ZIndex: int(p.ZIndex)}
}

func positionFrom(p nui.Position) Position {
	return Position{Type: PositionType(p.Type), Left: Length(p.Left), Top: Length(p.Top), Right: Length(p.Right), Bottom: Length(p.Bottom), ZIndex: int32(p.ZIndex)}
}

func (o Outline) nago() nui.Outline {
	return nui.Outline{Style: nui.OutlineStyle(o.Style), Width: int(o.Width), Offset: int(o.Offset), Color: nui.Color(o.Color)}
}

func outlineFrom(o nui.Outline) Outline {
	return Outline{Style: OutlineStyle(o.Style), Width: int32(o.Width), Offset: int32(o.Offset), Color: Color(o.Color)}
}

func (t Transformation) nago() nui.Transformation {
	return nui.Transformation{TranslateX: nui.Length(t.TranslateX), TranslateY: nui.Length(t.TranslateY), TranslateZ: nui.Length(t.TranslateZ), ScaleX: t.ScaleX, ScaleY: t.ScaleY, ScaleZ: t.ScaleZ, RotateZ: t.RotateZ}
}

func transformationFrom(t nui.Transformation) Transformation {
	return Transformation{TranslateX: Length(t.TranslateX), TranslateY: Length(t.TranslateY), TranslateZ: Length(t.TranslateZ), ScaleX: t.ScaleX, ScaleY: t.ScaleY, ScaleZ: t.ScaleZ, RotateZ: t.RotateZ}
}
