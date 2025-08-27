// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"fmt"

	"go.wdy.de/nago/presentation/proto"
)

// Alignment is specified as follows:
//
//	┌─TopLeading───────────Top─────────TopTrailing─┐
//	│                                              │
//	│                                              │
//	│                                              │
//	│                                              │
//	│                                              │
//	│                                              │
//	│                                              │
//	│ Leading            Center            Trailing│
//	│                                              │
//	│                                              │
//	│                                              │
//	│                                              │
//	│                                              │
//	│                                              │
//	│                                              │
//	└BottomLeading───────Bottom──────BottomTrailing┘
//
// Alignment is a layout component (Alignment).
// It defines the positioning of child views within a container,
// such as top, bottom, leading, trailing, or centered.
// An empty Alignment defaults to Center ("c") if not specified.
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Alignment uint

func (a Alignment) ora() proto.Alignment {
	return proto.Alignment(a)
}

func Alignments() []Alignment {
	return []Alignment{
		Top,
		Center,
		Bottom,
		Leading,
		Trailing,
		TopLeading,
		TopTrailing,
		BottomLeading,
		BottomTrailing,
	}
}

func (a Alignment) String() string {
	switch a {
	case Top:
		return "top"
	case Bottom:
		return "bottom"
	case Center:
		return "center"
	case Leading:
		return "leading"
	case Trailing:
		return "trailing"
	case TopLeading:
		return "top-leading"
	case BottomLeading:
		return "bottom-leading"
	case TopTrailing:
		return "top-trailing"
	case BottomTrailing:
		return "bottom-trailing"
	case Stretch:
		return "stretch"
	default:
		return fmt.Sprintf("%d", a)
	}
}

const (
	Center         Alignment = Alignment(proto.Center)
	Top            Alignment = Alignment(proto.Top)
	Bottom         Alignment = Alignment(proto.Bottom)
	Leading        Alignment = Alignment(proto.Leading)
	Trailing       Alignment = Alignment(proto.Trailing)
	TopLeading     Alignment = Alignment(proto.TopLeading)
	TopTrailing    Alignment = Alignment(proto.TopTrailing)
	BottomLeading  Alignment = Alignment(proto.BottomLeading)
	BottomTrailing Alignment = Alignment(proto.BottomTrailing)
	Stretch        Alignment = Alignment(proto.Stretch)
)
