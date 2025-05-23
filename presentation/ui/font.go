// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import "go.wdy.de/nago/presentation/proto"

type Font struct {
	// Name of the font or family name as fallback. Extra fallback declarations are unspecified and must be comma
	// separated.
	Name string

	// Size of the font
	Size Length

	Style FontStyle

	Weight FontWeight
}

func (f Font) ora() proto.Font {
	return proto.Font{
		Name:   proto.Str(f.Name),
		Size:   proto.Length(f.Size),
		Style:  proto.FontStyle(f.Style),
		Weight: proto.FontWeight(f.Weight),
	}
}

var (
	Title = Font{
		Size:   "1.5rem",
		Weight: BoldFontWeight,
	}

	SubTitle = Font{
		Size:   "1rem",
		Weight: BoldFontWeight,
	}

	Large = Font{
		Size:   "1.5rem",
		Weight: NormalFontWeight,
	}

	Small = Font{
		Size:   "0.75rem",
		Weight: NormalFontWeight,
	}

	Monospace = Font{
		Name: "monospace",
	}
)

type FontStyle uint

const (
	ItalicFontStyle FontStyle = FontStyle(proto.ItalicFontStyle)
	NormalFontStyle           = FontStyle(proto.NormalFontStyle)
)

type FontWeight int

const (
	NormalFontWeight FontWeight = 400
	BoldFontWeight   FontWeight = 700
)
