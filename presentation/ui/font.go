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
	Name FontName

	// Size of the font
	Size Length

	Style FontStyle

	Weight FontWeight

	LineHeight LineHeight
}

func (f Font) ora() proto.Font {
	return proto.Font{
		Name:       proto.Str(f.Name),
		Size:       proto.Length(f.Size),
		Style:      proto.FontStyle(f.Style),
		Weight:     proto.FontWeight(f.Weight),
		LineHeight: proto.LineHeight(f.LineHeight),
	}
}

var (
	// DisplayLarge is primarily meant for large stylistic text like in hero or header elements
	DisplayLarge = Font{
		Name:       DefaultFontName,
		Size:       "3.5625rem",
		Weight:     DisplayAndLabelFontWeight,
		LineHeight: "3.875rem",
	}

	// DisplayMedium is primarily meant for large stylistic text like in hero or header elements
	DisplayMedium = Font{
		Name:       DefaultFontName,
		Size:       "2.8125rem",
		Weight:     DisplayAndLabelFontWeight,
		LineHeight: "3.25rem",
	}

	// DisplaySmall is primarily meant for large stylistic text like in hero or header elements
	DisplaySmall = Font{
		Name:       DefaultFontName,
		Size:       "2.25rem",
		Weight:     DisplayAndLabelFontWeight,
		LineHeight: "2.75rem",
	}

	// HeadlineLarge is primarily meant for large headlines above sections and modules
	HeadlineLarge = Font{
		Name:       DefaultFontName,
		Size:       "2.5rem",
		Weight:     HeadlineAndTitleFontWeight,
		LineHeight: "3rem",
	}

	// HeadlineMedium is primarily meant for medium headlines above sections and modules
	HeadlineMedium = Font{
		Name:       DefaultFontName,
		Size:       "1.875rem",
		Weight:     HeadlineAndTitleFontWeight,
		LineHeight: "2.4375rem",
	}

	// HeadlineSmall is primarily meant for small headlines above sections and modules
	HeadlineSmall = Font{
		Name:       DefaultFontName,
		Size:       "1.5rem",
		Weight:     HeadlineAndTitleFontWeight,
		LineHeight: "1.9375rem",
	}

	// TitleLarge has the same size as BodyLarge but with bold font and smaller LineHeight
	TitleLarge = Font{
		Name:       DefaultFontName,
		Size:       "1rem",
		Weight:     HeadlineAndTitleFontWeight,
		LineHeight: "1.375rem",
	}

	// TitleMedium is the same as BodyMedium but with bold font
	TitleMedium = Font{
		Name:       DefaultFontName,
		Size:       "0.875rem",
		Weight:     HeadlineAndTitleFontWeight,
		LineHeight: "1.25rem",
	}

	// TitleSmall is the same as BodySmall but with bold font
	TitleSmall = Font{
		Name:       DefaultFontName,
		Size:       "0.75rem",
		Weight:     HeadlineAndTitleFontWeight,
		LineHeight: "1rem",
	}

	// BodyLarge is primarily meant for the normal document text.
	//
	// This is the default font if no font is specified.
	BodyLarge = Font{
		Name:       DefaultFontName,
		Size:       "1rem",
		Weight:     BodyFontWeight,
		LineHeight: "1.625rem",
	}

	// BodyMedium is primarily meant for the normal document text
	BodyMedium = Font{
		Name:       DefaultFontName,
		Size:       "0.875rem",
		Weight:     BodyFontWeight,
		LineHeight: "1.25rem",
	}

	// BodySmall is primarily meant for the normal document text
	BodySmall = Font{
		Name:       DefaultFontName,
		Size:       "0.75rem",
		Weight:     BodyFontWeight,
		LineHeight: "1rem",
	}

	// LabelLarge is primarily meant for texts in buttons, menus, ...
	LabelLarge = Font{
		Name:       DefaultFontName,
		Size:       "1.0625rem",
		Weight:     DisplayAndLabelFontWeight,
		LineHeight: "1.0625rem",
	}

	// LabelMedium is primarily meant for texts in buttons, menus, ...
	LabelMedium = Font{
		Name:       DefaultFontName,
		Size:       "0.875rem",
		Weight:     DisplayAndLabelFontWeight,
		LineHeight: "0.875rem",
	}

	// LabelSmall is primarily meant for texts in buttons, menus, ...
	LabelSmall = Font{
		Name:       DefaultFontName,
		Size:       "0.75rem",
		Weight:     DisplayAndLabelFontWeight,
		LineHeight: "0.8125rem",
	}

	// MonoLarge is primarily meant for monospaced font e.g. in code blocks
	MonoLarge = Font{
		Name:   MonoFontName,
		Size:   "1.5rem",
		Weight: MonoFontWeight,
	}

	// MonoMedium is primarily meant for monospaced font e.g. in code blocks
	MonoMedium = Font{
		Name:   MonoFontName,
		Size:   "1rem",
		Weight: MonoFontWeight,
	}

	// MonoSmall is primarily meant for monospaced font e.g. in code blocks
	MonoSmall = Font{
		Name:   MonoFontName,
		Size:   "0.75rem",
		Weight: MonoFontWeight,
	}

	// MonoBoldLarge is primarily meant for monospaced font e.g. in code blocks
	MonoBoldLarge = Font{
		Name:   MonoFontName,
		Size:   "1.5rem",
		Weight: MonoBoldFontWeight,
	}

	// MonoBoldMedium is primarily meant for monospaced font e.g. in code blocks
	MonoBoldMedium = Font{
		Name:   MonoFontName,
		Size:   "1rem",
		Weight: MonoBoldFontWeight,
	}

	// MonoBoldSmall is primarily meant for monospaced font e.g. in code blocks
	MonoBoldSmall = Font{
		Name:   MonoFontName,
		Size:   "0.75rem",
		Weight: MonoBoldFontWeight,
	}

	// MonoItalicLarge is primarily meant for monospaced font e.g. in code blocks
	MonoItalicLarge = Font{
		Name:   MonoFontName,
		Size:   "1.5rem",
		Weight: MonoItalicFontWeight,
	}

	// MonoItalicMedium is primarily meant for monospaced font e.g. in code blocks
	MonoItalicMedium = Font{
		Name:   MonoFontName,
		Size:   "1rem",
		Weight: MonoItalicFontWeight,
	}

	// MonoItalicSmall is primarily meant for monospaced font e.g. in code blocks
	MonoItalicSmall = Font{
		Name:   MonoFontName,
		Size:   "0.75rem",
		Weight: MonoItalicFontWeight,
	}

	// Deprecated: Title is a legacy font variation and should not be used in future projects.
	// Use TitleLarge, TitleMedium or TitleSmall instead.
	Title = Font{
		Size:   "1.5rem",
		Weight: HeadlineAndTitleFontWeight,
	}

	// Deprecated: SubTitle is a legacy font variation and should not be used in future projects.
	// Use TitleLarge, TitleMedium or TitleSmall instead.
	SubTitle = Font{
		Size:   "1rem",
		Weight: HeadlineAndTitleFontWeight,
	}

	// Deprecated: Large is a legacy font variation and should not be used in future projects.
	// Use other variations like BodyLarge, HeadlineLarge or DisplayLarge instead.
	Large = Font{
		Size:   "1.5rem",
		Weight: BodyFontWeight,
	}

	// Deprecated: Small is a legacy font variation and should not be used in future projects.
	// Use other variations like BodySmall, HeadlineSmall or DisplaySmall instead.
	Small = Font{
		Size:   "0.75rem",
		Weight: BodyFontWeight,
	}

	// Deprecated: Monospace is a legacy font variation and should not be used in future projects.
	// Use MonoLarge, MonoMedium, MonoSmall, MonoBoldLarge, MonoBoldMedium, MonoBoldSmall, MonoItalicLarge, MonoItalicMedium or MonoItalicSmall instead.
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
	BodyFontWeight             FontWeight = 400
	MonoFontWeight             FontWeight = 400
	MonoItalicFontWeight       FontWeight = 400
	DisplayAndLabelFontWeight  FontWeight = 600
	HeadlineAndTitleFontWeight FontWeight = 700
	MonoBoldFontWeight         FontWeight = 700
)

type FontName string

const (
	DefaultFontName FontName = "InterVariable"
	MonoFontName    FontName = "IBM Plex Mono"
)

type LineHeight string
