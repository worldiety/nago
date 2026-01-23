// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"fmt"
	"strings"

	"github.com/worldiety/material-color-utilities/hct"
	"go.wdy.de/nago/pkg/xcolor"
	"go.wdy.de/nago/presentation/proto"
)

// Color specifies either a hex color like #rrggbb or #rrggbbaa or an internal custom color name.
// Optional opacity values of custom color names will be attached to the color name, separated by the `/` character,
// e.g. `M8/128` describes the custom color `M8` with opacity `128` (0-255)
type Color string

func (c Color) ora() proto.Color {
	return proto.Color(c)
}

func (c Color) isAbsolute() bool {
	return strings.HasPrefix(string(c), `#`)
}

// WithTransparency updates the alpha value part of the color (0-100), where 25% transparent means 75% opaque.
func (c Color) WithTransparency(a int8) Color {
	// recalculate into 0-255 and invert
	ai := 255 - int(min(max(float64(a)/100*255, 0), 255))

	if c.isAbsolute() {
		if len(c) == 9 {
			co := c[:len(c)-2]

			return Color(fmt.Sprintf("%s%02x", string(co), ai))
		}
	}

	return Color(fmt.Sprintf("%s/%d", string(c), ai))
}

// WithoutTransparency updates the removes the alpha value of the color.
func (c Color) WithoutTransparency() Color {
	if c.isAbsolute() {
		if len(c) == 9 {
			return c[:len(c)-2]
		}

		return c
	}

	return Color(strings.Split(string(c), "/")[0])
}

// WithChromaAndTone applies the given chroma and tone values on the actual hue value using the HCT colorspace.
func (c Color) WithChromaAndTone(chroma float64, tone float64) (Color, error) {
	cl, err := xcolor.ParseHex(string(c))
	if err != nil {
		return c, err
	}

	v := hct.FromInt(cl.Int())
	v.SetChroma(chroma)
	v.SetTone(tone)
	return Color(xcolor.Hex(xcolor.RGBA8FromInt(v.ToInt()))), nil
}

const (
	// M0 is the source main color variable name.
	M0 Color = "M0"

	// M1 is a variable name usually used for the background.
	M1 Color = "M1"

	// M2 is a variable name usually used as the background for first level container
	M2 Color = "M2"

	// M3 is a variable name usually used for a card bottom area.
	M3 Color = "M3"

	// M4 is a variable name usually used for a card body area.
	M4 Color = "M4"

	// M5 is a variable name usually used for Line / Dot on SC.
	M5 Color = "M5"

	// M6 is a variable name usually used for hovered containers.
	M6 Color = "M6"

	// M7 is a variable name usually used for Text or muted icons.
	M7 Color = "M7"

	// M8 is a variable name usually used for Text or icons.
	M8 Color = "M8"

	// M9 is a variable name usually used for card Top area.
	M9 Color = "M9"

	// A0 is the source accent color name.
	A0 Color = "A0"

	// A1 is a variable name usually used for card progress bars, H2 or Borders.
	A1 Color = "A1"

	// I0 is the source interactive color name.
	I0 Color = "I0"

	// I1 is a variable name usually used for buttons.
	I1 Color = "I1"

	// SE0 is the source error color variable name.
	SE0 Color = "SE0"

	// SW0 is the source warning color variable name.
	SW0 Color = "SW0"

	// SG0 is the source good color variable name.
	SG0 Color = "SG0"

	// SV0 is the source Informative color variable name.
	SV0 Color = "SV0"

	// SI0 is the source disabled input color variable name.
	SI0 Color = "SI0"

	// ST0 is the source disabled text color variable name.
	ST0 Color = "ST0"
)

// additional alias names for base colors
const (
	// ColorCardBody represents the variable name which contains the conventional card body color derived from
	// the main color.
	ColorCardBody = M4

	ColorCardTop = M9

	ColorCardFooter = M3

	ColorContainer = M6

	ColorBackground = M1

	// ColorAccent represents the variable name containing the exact accent color.
	ColorAccent = A0

	// ColorInputBorder represents the variable name which refers to the color of the border for input elements like
	// a text field.
	ColorInputBorder = M8

	ColorText = M8

	// ColorLine represents the variable name containing the conventional color for a line derived from the main color.
	ColorLine = M5

	// ColorError represents the variable name which refers to the error color value.
	ColorError = SE0

	ColorBlack = "#000000"
	ColorWhite = "#ffffff"

	// ColorIcons or default text.
	ColorIcons = M8

	ColorIconsMuted = M7

	ColorInteractive   = I0
	ColorSemanticGood  = SG0
	ColorSemanticWarn  = SW0
	ColorSemanticError = SE0

	ColorBannerErrorBackground = "CBEB"
	ColorBannerErrorText       = "CBET"
	ColorBannerInfoBackground  = "CBIB"
	ColorBannerInfoText        = "CBIT"
)
