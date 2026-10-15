// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//gostrict:header
package ui

import nui "go.wdy.de/nago/presentation/ui"

// Font is the font of a text.
type Font struct {
	// Name is the name of the font or family.
	Name   FontName
	Size   Length
	Style  FontStyle
	Weight FontWeight
	// LineHeight is a length of CSS.
	LineHeight LineHeight
}

// FontStyle is normal or italic.
type FontStyle uint32

const (
	NormalFontStyle FontStyle = 0
	ItalicFontStyle FontStyle = 1
)

// FontWeight is the weight of a font, e.g. 400 for normal and 700 for bold.
type FontWeight int32

const (
	BodyFontWeight             FontWeight = 400
	MonoFontWeight             FontWeight = 400
	MonoItalicFontWeight       FontWeight = 400
	DisplayAndLabelFontWeight  FontWeight = 600
	HeadlineAndTitleFontWeight FontWeight = 700
	MonoBoldFontWeight         FontWeight = 700
)

// FontName is the name of a font.
type FontName string

const (
	DefaultFontName FontName = "InterVariable"
	MonoFontName    FontName = "IBM Plex Mono"
)

// LineHeight is the height of a line of CSS.
type LineHeight string

// DisplayLarge is for large stylistic text like in a hero or a header.
//
//gostrict:js ({ Name: "InterVariable", Size: "3.5625rem", Style: 0, Weight: 600, LineHeight: "3.875rem" })
var DisplayLarge Font = fontFrom(nui.DisplayLarge)

// DisplayMedium is primarily meant for large stylistic text like in hero or header elements.
//
//gostrict:js ({ Name: "InterVariable", Size: "2.8125rem", Style: 0, Weight: 600, LineHeight: "3.25rem" })
var DisplayMedium Font = fontFrom(nui.DisplayMedium)

// DisplaySmall is primarily meant for large stylistic text like in hero or header elements.
//
//gostrict:js ({ Name: "InterVariable", Size: "2.25rem", Style: 0, Weight: 600, LineHeight: "2.75rem" })
var DisplaySmall Font = fontFrom(nui.DisplaySmall)

// HeadlineLarge is primarily meant for large headlines above sections and modules.
//
//gostrict:js ({ Name: "InterVariable", Size: "2.5rem", Style: 0, Weight: 700, LineHeight: "3rem" })
var HeadlineLarge Font = fontFrom(nui.HeadlineLarge)

// HeadlineMedium is primarily meant for medium headlines above sections and modules.
//
//gostrict:js ({ Name: "InterVariable", Size: "1.875rem", Style: 0, Weight: 700, LineHeight: "2.4375rem" })
var HeadlineMedium Font = fontFrom(nui.HeadlineMedium)

// HeadlineSmall is for the headline of a section.
//
//gostrict:js ({ Name: "InterVariable", Size: "1.5rem", Style: 0, Weight: 700, LineHeight: "1.9375rem" })
var HeadlineSmall Font = fontFrom(nui.HeadlineSmall)

// TitleLarge has the same size as BodyLarge but with bold font and smaller LineHeight.
//
//gostrict:js ({ Name: "InterVariable", Size: "1rem", Style: 0, Weight: 700, LineHeight: "1.375rem" })
var TitleLarge Font = fontFrom(nui.TitleLarge)

// TitleMedium is the same as BodyMedium but with bold font.
//
//gostrict:js ({ Name: "InterVariable", Size: "0.875rem", Style: 0, Weight: 700, LineHeight: "1.25rem" })
var TitleMedium Font = fontFrom(nui.TitleMedium)

// TitleSmall is the same as BodySmall but with bold font.
//
//gostrict:js ({ Name: "InterVariable", Size: "0.75rem", Style: 0, Weight: 700, LineHeight: "1rem" })
var TitleSmall Font = fontFrom(nui.TitleSmall)

// BodyLarge is primarily meant for the normal document text. It is the default font, if no font is specified.
//
//gostrict:js ({ Name: "InterVariable", Size: "1rem", Style: 0, Weight: 400, LineHeight: "1.625rem" })
var BodyLarge Font = fontFrom(nui.BodyLarge)

// BodyMedium is the font of a text.
//
//gostrict:js ({ Name: "InterVariable", Size: "0.875rem", Style: 0, Weight: 400, LineHeight: "1.25rem" })
var BodyMedium Font = fontFrom(nui.BodyMedium)

// BodySmall is primarily meant for the normal document text.
//
//gostrict:js ({ Name: "InterVariable", Size: "0.75rem", Style: 0, Weight: 400, LineHeight: "1rem" })
var BodySmall Font = fontFrom(nui.BodySmall)

// LabelLarge is primarily meant for texts in buttons, menus, ...
//
//gostrict:js ({ Name: "InterVariable", Size: "1rem", Style: 0, Weight: 600, LineHeight: "1.0625rem" })
var LabelLarge Font = fontFrom(nui.LabelLarge)

// LabelMedium is primarily meant for texts in buttons, menus, ...
//
//gostrict:js ({ Name: "InterVariable", Size: "0.875rem", Style: 0, Weight: 600, LineHeight: "0.9375rem" })
var LabelMedium Font = fontFrom(nui.LabelMedium)

// LabelSmall is primarily meant for texts in buttons, menus, ...
//
//gostrict:js ({ Name: "InterVariable", Size: "0.75rem", Style: 0, Weight: 600, LineHeight: "0.8125rem" })
var LabelSmall Font = fontFrom(nui.LabelSmall)

// MonoLarge is primarily meant for monospaced font e.g. in code blocks.
//
//gostrict:js ({ Name: "IBM Plex Mono", Size: "1.5rem", Style: 0, Weight: 400, LineHeight: "" })
var MonoLarge Font = fontFrom(nui.MonoLarge)

// MonoMedium is primarily meant for monospaced font e.g. in code blocks.
//
//gostrict:js ({ Name: "IBM Plex Mono", Size: "1rem", Style: 0, Weight: 400, LineHeight: "" })
var MonoMedium Font = fontFrom(nui.MonoMedium)

// MonoSmall is primarily meant for monospaced font e.g. in code blocks.
//
//gostrict:js ({ Name: "IBM Plex Mono", Size: "0.75rem", Style: 0, Weight: 400, LineHeight: "" })
var MonoSmall Font = fontFrom(nui.MonoSmall)

// MonoBoldLarge is primarily meant for monospaced font e.g. in code blocks.
//
//gostrict:js ({ Name: "IBM Plex Mono", Size: "1.5rem", Style: 0, Weight: 700, LineHeight: "" })
var MonoBoldLarge Font = fontFrom(nui.MonoBoldLarge)

// MonoBoldMedium is primarily meant for monospaced font e.g. in code blocks.
//
//gostrict:js ({ Name: "IBM Plex Mono", Size: "1rem", Style: 0, Weight: 700, LineHeight: "" })
var MonoBoldMedium Font = fontFrom(nui.MonoBoldMedium)

// MonoBoldSmall is primarily meant for monospaced font e.g. in code blocks.
//
//gostrict:js ({ Name: "IBM Plex Mono", Size: "0.75rem", Style: 0, Weight: 700, LineHeight: "" })
var MonoBoldSmall Font = fontFrom(nui.MonoBoldSmall)

// MonoItalicLarge is primarily meant for monospaced font e.g. in code blocks.
//
//gostrict:js ({ Name: "IBM Plex Mono", Size: "1.5rem", Style: 1, Weight: 400, LineHeight: "" })
var MonoItalicLarge Font = fontFrom(nui.MonoItalicLarge)

// MonoItalicMedium is primarily meant for monospaced font e.g. in code blocks.
//
//gostrict:js ({ Name: "IBM Plex Mono", Size: "1rem", Style: 1, Weight: 400, LineHeight: "" })
var MonoItalicMedium Font = fontFrom(nui.MonoItalicMedium)

// MonoItalicSmall is primarily meant for monospaced font e.g. in code blocks.
//
//gostrict:js ({ Name: "IBM Plex Mono", Size: "0.75rem", Style: 1, Weight: 400, LineHeight: "" })
var MonoItalicSmall Font = fontFrom(nui.MonoItalicSmall)

// Title is a legacy font variation and should not be used in future projects. Use TitleLarge, TitleMedium or
// TitleSmall instead.
//
//gostrict:js ({ Name: "", Size: "1.5rem", Style: 0, Weight: 700, LineHeight: "" })
var Title Font = fontFrom(nui.Title)

// SubTitle is a legacy font variation and should not be used in future projects. Use TitleLarge, TitleMedium or
// TitleSmall instead.
//
//gostrict:js ({ Name: "", Size: "1rem", Style: 0, Weight: 700, LineHeight: "" })
var SubTitle Font = fontFrom(nui.SubTitle)

// Large is a legacy font variation and should not be used in future projects. Use other variations like BodyLarge,
// HeadlineLarge or DisplayLarge instead.
//
//gostrict:js ({ Name: "", Size: "1.5rem", Style: 0, Weight: 400, LineHeight: "" })
var Large Font = fontFrom(nui.Large)

// Small is a legacy font variation and should not be used in future projects. Use other variations like BodySmall,
// HeadlineSmall or DisplaySmall instead.
//
//gostrict:js ({ Name: "", Size: "0.75rem", Style: 0, Weight: 400, LineHeight: "" })
var Small Font = fontFrom(nui.Small)

// Monospace is a legacy font variation and should not be used in future projects. Use MonoLarge, MonoMedium,
// MonoSmall, MonoBoldLarge, MonoBoldMedium, MonoBoldSmall, MonoItalicLarge, MonoItalicMedium or MonoItalicSmall
// instead.
//
//gostrict:js ({ Name: "monospace", Size: "", Style: 0, Weight: 0, LineHeight: "" })
var Monospace Font = fontFrom(nui.Monospace)
